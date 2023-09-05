APP_VERSION = 0.1

DB_USER     = admin
DB_PASSWORD = admin
DB_NAME     = redditsoccergoals
DB_PORT     = 5432

clean:
	rm -r bin/*

build-all:
	go build -v ./...

# poller commands ---------------------------------------------------------------------------------------------------------
run-poller:
	go run ./cmd/poller/...

run-poller-docker:
	docker run --env-file .env.docker top90-poller-v0.1

build-poller:
	cd cmd/poller && go build -o ../../bin/goal_poller
	docker build -f Dockerfile.poller -t top90-poller-v${APP_VERSION} .

build-poller-linux:
	cd cmd/poller && GOOS=linux GOARCH=amd64 go build -o ../../bin/goal_poller_linux
	docker build -f Dockerfile.poller --platform=linux/amd64 -t top90-poller-v${APP_VERSION} .

deploy-poller: build-poller-linux
	scp -i keys/defaultec2.pem bin/goal_poller_linux ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com:~/.

# apifootball commands ----------------------------------------------------------------------------------------------------

build-apifootball-ingest-linux:
	cd cmd/apifootball_ingest && GOOS=linux GOARCH=amd64 go build -o ../../bin/apifootball_ingest_linux

deploy-apifootball-ingest: build-apifootball-ingest-linux
	scp -i keys/defaultec2.pem bin/apifootball_ingest_linux ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com:~/.

# server commands ---------------------------------------------------------------------------------------------------------
run-server:
	go run ./cmd/server/...

build-server:
	cd cmd/server && go build -o ../../bin/server
	docker build --platform=linux/amd64 -t top90-server-v${APP_VERSION} .

save-server-image:
	docker save -o ./bin/top90-server-v${APP_VERSION}.tar top90-server-v${APP_VERSION}

copy-server-image-to-ec2:
	scp -i keys/defaultec2.pem bin/top90-server-v${APP_VERSION}.tar ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com:~/.

deploy-server: build-server save-server-image copy-server-image-to-ec2
	ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com docker load --input top90-server-v${APP_VERSION}.tar
	-ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com docker stop top90-server
	ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com docker container prune -f
	ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com docker image prune -f
	ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com docker run -p 7171:7171 -d --restart unless-stopped --name top90-server top90-server-v${APP_VERSION}

# playground commands -----------------------------------------------------------------------------------------------------
run-playground:
	go run ./cmd/playground/...

# apifootball ingest commands --------------------------------------------------------------------------------------------------
run-apifootball-ingest:
	go run ./cmd/apifootball_ingest/... ${TYPE}

# db migration commands ---------------------------------------------------------------------------------------------------
migrate-up:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable" -path internal/db/migrations up

migrate-down:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable" -path internal/db/migrations down

migrate-rollback:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable" -path internal/db/migrations down 1

# tunneling / ec2 commands ------------------------------------------------------------------------------------------------
ssh-ec2:
	ssh -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com

# -N do not execute remote command | -v for verbose mode | -f to make it go to the background
tunnel-prod-db:
	ssh -i keys/defaultec2.pem -N -L 5433:reddit-soccer-goals.cxdhgbr8e3pn.us-east-1.rds.amazonaws.com:5432 ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com -v

get-poller-logs:
	scp -i keys/defaultec2.pem ec2-user@ec2-18-232-86-153.compute-1.amazonaws.com:~/goal_poller_output.txt .

# utility commands --------------------------------------------------------------------------------------------------------
create-s3-bucket:
	aws --endpoint-url=http://localhost:4566 s3 mb s3://reddit-soccer-goals

ingest-apifootball:
	make run-apifootball-ingest TYPE=leagues
	make run-apifootball-ingest TYPE=teams
	make run-apifootball-ingest TYPE=fixtures

clear-tmp:
	rm -rf tmp
	mkdir tmp

seed: clear-tmp migrate-down migrate-up create-s3-bucket build-poller run-poller-docker