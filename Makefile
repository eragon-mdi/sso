include .env
export

CONTAINER_DB_POSTGRES_NAME=fast-postgres
CONTAINER_DB_REDIS_NAME=fast-redis

# ======= RSA =======
gen-rsa-pair:
	mkdir -p ./secrets/
	openssl genrsa -out ./secrets/private.pem 2048
	openssl rsa -in ./secrets/private.pem -pubout -out ./secrets/public.pem
	
# ======= BUILD =======
restart-quiet: down start-quiet
start-quiet:
	docker compose build && \
	docker compose up -d && \
	clear && \
	docker ps -a
down:
	docker compose down

# ======= DEV FAST BUILD =======
dev-restart: dev-down dev-up
dev-up: run-postgres run-redis wait migrate-up run-app
dev-down: rm-redis rm-postgres
##	
run-postgres:
	docker network create bridge_app || true
	docker run \
		--name $(CONTAINER_DB_POSTGRES_NAME) \
		--network=bridge_app \
		-e POSTGRES_USER=$(STORAGES_POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(STORAGES_POSTGRES_PASS) \
		-e POSTGRES_DB=$(STORAGES_POSTGRES_NAME) \
		-p $(STORAGES_POSTGRES_PORT):$(STORAGES_POSTGRES_PORT) \
		-d \
		postgres:latest
rm-postgres:
	docker rm -f $(CONTAINER_DB_POSTGRES_NAME) || true
	docker network rm bridge_app 

run-redis:
	docker run \
	  --name $(CONTAINER_DB_REDIS_NAME) \
	  --network=bridge_app \
	  -e REDIS_PASSWORD=$(STORAGES_REDIS_PASS) \
	  -p $(STORAGES_REDIS_PORT):6379 \
	  -d \
	  redis:8.2.1-alpine \
	  redis-server --requirepass $(STORAGES_REDIS_PASS)
rm-redis:
	docker rm -f $(CONTAINER_DB_REDIS_NAME) || true

migrate-up:
	migrate -path ./migrations/sql -database "postgres://$(STORAGES_POSTGRES_USER):$(STORAGES_POSTGRES_PASS)@$(STORAGES_POSTGRES_HOST):$(STORAGES_POSTGRES_PORT)/$(STORAGES_POSTGRES_NAME)?sslmode=$(STORAGES_POSTGRES_SSLM)" up
##	 
#-tags=dev
run-app:
	go run cmd/sso/main.go
##	clear-port: 					# if don't correct close app
##		lsof -ti :$(SERVER_PORT)
##		kill -9 $$(lsof -ti :$(SERVER_PORT))
##	
##	
# ======= DEV TOOLS =======
# Need new migrate files
migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: укажи имя миграции через 'name=...'" && exit 1; \
	fi
	migrate create -ext sql -dir ./migrations/sql $(name)
# p.s. 
#curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz
#sudo mv migrate /usr/local/bin
##	
lint:
	golangci-lint run ./cmd/... ./internal/...
##	
##	swag:
##		swag init -g ./cmd/url-short/main.go
##	
gen-mocks:
	go generate ./internal/...

gen-base-tests-transport:
	gotests -w -all ./internal/transport/http2/grpc/sso/auth/auth.go

gen-base-tests-service:
	gotests -w -all \
		./internal/service/sso/auth/auth.go \
		./internal/service/sso/auth/common.go \
		./internal/service/sso/permission/permission.go
##	
go-tests-coverage:
	mkdir -p ./tmp
	go test \
		-short \
		-count=1 \
		-race \
		-coverprofile=./tmp/coverage.out \
		./internal/service/... \
		./internal/transport/...
	go tool cover \
		-html=./tmp/coverage.out \
		-o ./tmp/coverage.html
	firefox tmp/coverage.html
	
go-tests-coverage-file:
	go test -short -count=1 -race -coverprofile=./tmp/coverage_file.out ./internal/service \
    	-coverpkg=./internal/service/sso/auth,./internal/service/sso/permission
	go tool cover -html=./tmp/coverage_file.out -o ./tmp/coverage_file.html
	firefox ./tmp/coverage_file.html


##	# ======= FUNCTIONAL TESTS =======
##	functional-auto: functional-auto-down functional-auto-up rm-postgres
##	
##	functional-auto-up: run-postgres wait migrate-up
##		go build -o ./tests/functional/app_bin ./cmd/url-short/main.go
##		go test -v ./tests/functional/ || true
##		rm ./tests/functional/app_bin
##	functional-auto-down: rm-postgres
##	
##	# common
wait:
	@sleep 2