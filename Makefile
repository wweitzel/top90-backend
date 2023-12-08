APP_VERSION = 0.1

# scraper --------------------------------------------------------------------------------------------------------
run-scraper:
	go run ./cmd/scraper

run-scraper-docker:
	docker run --rm --env-file .env.docker top90-scraper-v0.1

build-scraper:
	cd cmd/scraper && go build -o ../../bin/goal_scraper
	docker build -f Dockerfile.scraper -t top90-scraper-v${APP_VERSION} .

# api ------------------------------------------------------------------------------------------------------------
run-api:
	go run ./cmd/api

run-api-docker:
	docker run --rm -p 7171:7171 --env-file .env.docker top90-api-v0.1

build-api:
	cd cmd/api && go build -o ../../bin/api
	docker build -t top90-api-v${APP_VERSION} .

build-api-linux:
	cd cmd/api && GOOS=linux GOARCH=amd64 go build -o ../../bin/api_linux
	docker build --platform=linux/amd64 -t top90-api-v${APP_VERSION} .

# syncdata -------------------------------------------------------------------------------------------------------
run-syncdata:
	go run ./cmd/syncdata

# misc -----------------------------------------------------------------------------------------------------------
clean:
	rm -rfv bin/*
	rm -rfv tmp/*

migrate-down:
	migrate -database "postgres://admin:admin@localhost:5434/redditsoccergoals?sslmode=disable" -path internal/db/migrations down

migrate-up:
	migrate -database "postgres://admin:admin@localhost:5434/redditsoccergoals?sslmode=disable" -path internal/db/migrations up

seed:
	go run ./scripts/seed_local_db/...
	make build-scraper
	make run-scraper-docker