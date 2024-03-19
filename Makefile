PROJECTNAME := $(shell basename "$(PWD)")
include .env
export $(shell sed 's/=.*//' .env)

## service-up: Run the all components by deployment/compose.yaml
.PHONY: service-up
service-up:
	@docker-compose  -f ./deployment/compose.yaml --project-directory . up

## service-down: Docker-compose down
.PHONY: service-down
service-down:
	@docker-compose -f ./deployment/compose.yaml --project-directory . down

.PHONY: generate-image
generate-image:
	@docker build --tag web3-ecommerce:$(shell git rev-parse HEAD) -f ./build/Dockerfile .

.PHONY: dynamodb-up
dynamodb-up:
	@aws dynamodb create-table --cli-input-json file://deployment/dynamodb/create-table.json --endpoint-url http://localhost:8000