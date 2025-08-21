package configs

import (
	"log"
	"time"
)

type Config struct {
	Storages      Storages      `envconfig:"STORAGES" required:"true"`
	Servers       Servers       `envconfig:"SERVERS" required:"true"`
	Logger        Logger        `envconfig:"LOGGER" required:"true"`
	BussinesLogic BussinesLogic `envconfig:"BUSSINES_LOGIC" required:"true"`
}

// func init() {
func MustLoad() *Config {
	if err := loadCfg(singleConfig); err != nil {
		log.Fatalf("failed load cfgs on init stage: %v", err)
	}
	return singleConfig
}

var singleConfig = &Config{}

func Get() *Config {
	return singleConfig
}

type Storages struct {
	Postgres PsqlStore  `envconfig:"POSTGRES" required:"true"`
	Redis    RedisStore `envconfig:"REDIS" required:"true"`
}

type PsqlStore struct {
	HostF     string `envconfig:"HOST" required:"true"`
	PortF     string `envconfig:"PORT" default:"5432"`
	UserF     string `envconfig:"USER" required:"true"`
	PasswordF string `envconfig:"PASS" required:"true"`
	NameF     string `envconfig:"NAME" required:"true"`
	SSLmodeF  string `envconfig:"SSLM" default:"disable"`
}

type RedisStore struct {
	HostF     string `envconfig:"HOST" required:"true"`
	PortF     string `envconfig:"PORT" default:"5432"`
	PasswordF string `envconfig:"PASS" required:"true"`
	DBNumberF int    `envconfig:"DB_NUMBER" default:"0"`
}

type Servers struct {
	GRPC Server `envconfig:"GRPC"`
}

type Server struct {
	AddressF           string        `envconfig:"ADDR" required:"true"`
	PortF              string        `envconfig:"PORT" required:"true"`
	ReadTimeoutF       time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
	WriteTimeoutF      time.Duration `envconfig:"WRITE_TIMEOUT" default:"5s"`
	ReadHeaderTimeoutF time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"5s"`
	IdleTimeoutF       time.Duration `envconfig:"IDLE_TIMEOUT" default:"5s"`
}

type Logger struct {
	Level      string `envconfig:"LEVEL" default:"error"`
	Encoding   string `envconfig:"ENCODING" default:"json"`
	Output     string `envconfig:"OUTPUT" default:"stdout"`
	MessageKey string `envconfig:"MESSAGE_KEY" default:"message"`
}

type BussinesLogic struct {
	PassHasherCost       int           `envconfig:"PASS_HASHER_COST" required:"true"`
	TokenTTL             time.Duration `envconfig:"TOKEN_TTL" required:"true"`
	PathSecretPrivate    string        `envconfig:"PATH_SECRET_PRIVATE" required:"true"`
	PathSecretPublic     string        `envconfig:"PATH_SECRET_PUBLIC" required:"true"`
	SecretForTokerHasher string        `envconfig:"SECRET_FOR_TOKER_HASHER" required:"true"`
}
