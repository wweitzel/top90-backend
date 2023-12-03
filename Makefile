APP_VERSION = 0.1

# poller ---------------------------------------------------------------------------------------------------------
run-poller:
	go run ./cmd/poller/...

run-poller-docker:
	docker run --rm --env-file .env.docker top90-poller-v0.1

build-poller:
	cd cmd/poller && go build -o ../../bin/goal_poller
	docker build -f Dockerfile.poller -t top90-poller-v${APP_VERSION} .

build-poller-linux:
	cd cmd/poller && GOOS=linux GOARCH=amd64 go build -o ../../bin/goal_poller_linux
	docker build -f Dockerfile.poller --platform=linux/amd64 -t top90-poller-v${APP_VERSION} .

# api ------------------------------------------------------------------------------------------------------------
run-api:
	go run ./cmd/api/...

run-api-docker:
	docker run --rm -p 7171:7171 --env-file .env.docker top90-api-v0.1

build-api:
	cd cmd/api && go build -o ../../bin/api
	docker build -t top90-api-v${APP_VERSION} .

build-api-linux:
	cd cmd/api && GOOS=linux GOARCH=amd64 go build -o ../../bin/api_linux
	docker build --platform=linux/amd64 -t top90-api-v${APP_VERSION} .

# apifootball ----------------------------------------------------------------------------------------------------
run-apifootball-ingest:
	go run ./cmd/apifootball_ingest/... ${TYPE}

build-apifootball-ingest-linux:
	cd cmd/apifootball_ingest && GOOS=linux GOARCH=amd64 go build -o ../../bin/apifootball_ingest_linux

deploy-apifootball-ingest: build-apifootball-ingest-linux
	scp -i keys/defaultec2.pem bin/apifootball_ingest_linux ec2-user@ec2-35-171-182-157.compute-1.amazonaws.com:~/.

# misc -----------------------------------------------------------------------------------------------------------
clean:
	rm -rfv bin/*
	rm -rfv tmp/*

seed:
	go run ./scripts/seed_local_db/...
	make build-poller
	make run-poller-docker