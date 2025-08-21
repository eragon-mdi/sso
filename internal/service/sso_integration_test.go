package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	pgdriver "github.com/eragon-mdi/go-playground/storage/drivers/postgres"
	redisstore "github.com/eragon-mdi/go-playground/storage/nosql/redis"
	sqlstore "github.com/eragon-mdi/go-playground/storage/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/eragon-mdi/sso/internal/common/configs"
	"github.com/eragon-mdi/sso/internal/domain"
	"github.com/eragon-mdi/sso/internal/repository"
	"github.com/eragon-mdi/sso/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	globalPgC tc.Container
	globalRC  tc.Container

	globalSQL sqlstore.Storage
	globalRds redisstore.Storage

	rootPATH = rootPath()
)

// ---- верхняя тестовая обёртка, использующая globalSQL/globalRds ----
func TestAuth_FullFlow(t *testing.T) {
	ctx := context.Background()

	testStor := &testStorage{sql: globalSQL, redis: globalRds}
	repo := repository.New(testStor)
	bl := &configs.BussinesLogic{
		PassHasherCost:       10,
		TokenTTL:             72 * time.Hour,
		PathSecretPrivate:    filepath.Join(rootPATH, "secrets/private.pem"),
		PathSecretPublic:     filepath.Join(rootPATH, "secrets/public.pem"),
		SecretForTokerHasher: "test-secret",
	}
	svc, err := service.New(repo, bl)
	if err != nil {
		t.Fatalf("service.New failed: %v", err)
	}

	user := domain.User{Email: "fullflow@test.local", Password: "secret123"}

	// --- REGISTER ---
	t.Run("Register success", func(t *testing.T) {
		u, err := svc.Register(ctx, user)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
		if u.ID == "" {
			t.Fatal("expected user ID")
		}
	})

	t.Run("Register duplicate", func(t *testing.T) {
		_, err := svc.Register(ctx, user)
		if err == nil {
			t.Fatal("expected duplicate email error")
		}
	})

	// --- LOGIN ---
	var tok domain.Token
	t.Run("Login success", func(t *testing.T) {
		tok, err = svc.Login(ctx, domain.User{Email: user.Email, Password: "secret123"}, domain.NewDeviceCtx(1, 1))
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if tok.Access == "" || tok.Refresh == "" {
			t.Fatal("tokens not set")
		}
	})

	t.Run("Login wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, domain.User{Email: user.Email, Password: "wrongpass"}, domain.NewDeviceCtx(1, 1))
		if err == nil {
			t.Fatal("expected login failure")
		}
	})

	// --- REFRESH ---
	t.Run("Refresh token success", func(t *testing.T) {
		tok2, err := svc.Refresh(ctx, tok.Refresh, domain.NewDeviceCtx(1, 1))
		if err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}
		if tok2.Access == tok.Access || tok2.Refresh == tok.Refresh {
			t.Fatal("expected new tokens")
		}
	})

	t.Run("Refresh token invalid", func(t *testing.T) {
		_, err := svc.Refresh(ctx, "invalidtoken", domain.NewDeviceCtx(1, 2))
		if err == nil {
			t.Fatal("expected refresh failure")
		}
	})

	t.Run("Refresh token wrong context", func(t *testing.T) {
		_, err := svc.Refresh(ctx, tok.Refresh, domain.NewDeviceCtx(99, 99))
		if err == nil {
			t.Fatal("expected context mismatch failure")
		}
	})

	// --- LOGOUT ---
	t.Run("Logout success", func(t *testing.T) {
		err := svc.Logout(ctx, tok.Refresh, domain.NewDeviceCtx(1, 1))
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}
	})

	t.Run("Logout invalid token", func(t *testing.T) {
		err := svc.Logout(ctx, "notarealtoken", domain.NewDeviceCtx(1, 1))
		if err == nil {
			t.Fatal("expected logout failure")
		}
	})

	t.Run("Logout wrong context", func(t *testing.T) {
		err := svc.Logout(ctx, tok.Refresh, domain.NewDeviceCtx(99, 99))
		if err == nil {
			t.Fatal("expected logout context failure")
		}
	})
}

// /
// /
// /
func TestMain(m *testing.M) {
	// увеличим таймаут старта контейнеров
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1) стартуем контейнеры (GenericContainer с wait)
	pgC, pgHost, pgPort := startPostgres(ctx, testingLogger{})
	defer func() { _ = pgC.Terminate(ctx) }()

	rc, rHost, rPort := startRedis(ctx, testingLogger{})
	defer func() { _ = rc.Terminate(ctx) }()

	// --- Сформируем DSN для миграций ---
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"test", "test", pgHost, pgPort, "testdb",
	)

	// 2) запустим миграции (убедись, что путь корректен)
	if err := runMigrations(dsn, filepath.Join(rootPATH, "migrations/sql")); err != nil {
		// В локальной dev-среде иногда полезно логировать и выйти с кодом != 0
		fmt.Fprintf(os.Stderr, "migrations failed: %v\n", err)
		os.Exit(2)
	}

	// 3) создаём sqlstore.Storage и redisstore.Storage напрямую (ниже — как у storage.Conn)
	pgCfg := configs.PsqlStore{
		HostF:     pgHost,
		PortF:     fmt.Sprintf("%d", pgPort),
		UserF:     "test",
		PasswordF: "test",
		NameF:     "testdb",
		SSLmodeF:  "disable",
	}

	var err error
	globalSQL, err = sqlstore.Conn(context.Background(), pgCfg, pgdriver.Postgres{}, 30*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sqlstore.Conn failed: %v\n", err)
		os.Exit(2)
	}

	redisCfg := configs.RedisStore{
		HostF:     rHost,
		PortF:     fmt.Sprintf("%d", rPort),
		PasswordF: "",
		DBNumberF: 0,
	}
	globalRds, err = redisstore.Conn(context.Background(), redisCfg, 30*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redisstore.Conn failed: %v\n", err)
		os.Exit(2)
	}

	// Сохраним глобальные переменные для тестов
	globalPgC = pgC
	globalRC = rc

	code := m.Run()

	// cleanup
	_ = globalSQL.Close()
	_ = globalRds.Close()
	os.Exit(code)
}

// ---------------- helpers ----------------

type testingLogger struct{}

func (testingLogger) Logf(string, ...any) {}

func startPostgres(ctx context.Context, _ testingLogger) (tc.Container, string, int) {
	req := tc.ContainerRequest{
		Image: "postgres:17.5-alpine",
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		panic(fmt.Sprintf("start pg: %v", err))
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	return c, host, port.Int()
}

func startRedis(ctx context.Context, _ testingLogger) (tc.Container, string, int) {
	req := tc.ContainerRequest{
		Image:        "redis:8.2.1-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		panic(fmt.Sprintf("start redis: %v", err))
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "6379")
	return c, host, port.Int()
}

// ---------------- testStorage adapter ----------------
type testStorage struct {
	sql   sqlstore.Storage
	redis redisstore.Storage
}

func (t *testStorage) SQL() sqlstore.Storage     { return t.sql }
func (t *testStorage) Redis() redisstore.Storage { return t.redis }
func (t *testStorage) GracefulShutdown() error {
	_ = t.sql.Close()
	_ = t.redis.Close()
	return nil
}

func runMigrations(dsn, migrationsPath string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// Возвращает абсолютный путь к папке миграций для текущего сервиса
func rootPath() string {
	_, filename, _, _ := runtime.Caller(0) // путь к текущему файлу теста
	dir := filepath.Dir(filename)          // ./services/sso/internal/service
	return filepath.Join(dir, "../../")    // ./services/sso
}
