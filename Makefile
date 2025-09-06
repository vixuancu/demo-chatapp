include .env
export
MIGRATION_DIRS = internal/db/migrations
CONN_STRING = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
ENV_FILE = .env
# up and remove docker container
up-container:
	docker-compose up -d
remove-container:
	docker-compose down
# stop tạm thời container
stop-container:
	docker-compose stop
# khởi động lại container
restart-container:
	docker-compose restart
# generate sqlc
sqlc:
	sqlc generate
# create a new migration database (make migrate-create NAME=chatapp)
migrate-create:
	migrate create -ext sql -dir $(MIGRATION_DIRS) -seq $(NAME)
# run all pending migrations (make migrate-up)
migrate-up:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" up
# rollback the last migration (make migrate-down)
migrate-down:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" down 1
# rollback n versions
migrate-down-n:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" down $(N)
# trở về phiên bản cụ thể (make migrate-force VERSION=1)
migrate-force:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" force $(VERSION)
# drop everything (include schema and data) (make migrate-drop)
migrate-drop:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" drop
# apple migrations to a specific version (make migrate-goto VERSION=1) - goto này là chạy đến bao gồm phiên bản trước luôn
migrate-goto:
	migrate -path $(MIGRATION_DIRS) -database "$(CONN_STRING)" goto $(VERSION)
# run server
server:
	go run ./cmd/api
.PHONY: start-container remove-container stop-container start-container restart-container