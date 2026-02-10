include .env  
export

run:
	@go run ./cmd

migration:
	@goose create $(name) sql -s

goose-up:
	@goose up

goose-down:
	@goose down

sqlc:
	@sqlc generate
