DB_USER     = admin
DB_PASSWORD = admin
DB_NAME     = redditsoccergoals

VERSION = 0.1

clean:
	rm -r bin/*

build-all: build-poller

# poller commands ---------------------------------------------------------------------------------------------------------
run-poller:
	go run ./cmd/poller/...

build-poller-linux:
	cd cmd/poller && GOOS=linux go build -o ../../bin/goal_poller_linux

build-poller: build-poller-linux
	cd cmd/poller && go build -o ../../bin/goal_poller

deploy-poller: build-poller-linux
	scp -i keys/defaultec2.pem bin/goal_poller_linux ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com:~/.
	scp -i keys/defaultec2.pem .env ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com:~/.

# server commands ---------------------------------------------------------------------------------------------------------
run-server:
	go run ./cmd/server/...

build-server:
	cd cmd/server && go build -o ../../bin/server
	docker build -t top90-server-v${VERSION} .

save-server-image:
	docker save -o ./bin/top90-server-v${VERSION}.tar top90-server-v${VERSION}

copy-server-image-to-ec2:
	scp -i keys/defaultec2.pem bin/top90-server-v${VERSION}.tar ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com:~/.

deploy-server: build-server save-server-image copy-server-image-to-ec2
	ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com docker load --input top90-server-v${VERSION}.tar
	-ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com docker stop top90-server
	ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com docker container prune -f
	ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com docker image prune -f
	ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com docker run -p 7171:7171 -d --restart unless-stopped --name top90-server top90-server-v${VERSION}

# playground commands -----------------------------------------------------------------------------------------------------
run-playground:
	go run ./cmd/playground/...

# leagues ingest commands --------------------------------------------------------------------------------------------------
run-leagues-ingest:
	go run ./cmd/leagues_ingest/...

# db migration commands ---------------------------------------------------------------------------------------------------
migrate-up:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable" -path internal/db/migrations up

migrate-down:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable" -path internal/db/migrations down

migrate-rollback:
	migrate -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable" -path internal/db/migrations down 1

# tunneling / ec2 commands ------------------------------------------------------------------------------------------------
ssh-ec2:
	ssh -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com

# -N do not execute remote command | -v for verbose mode | -f to make it go to the background
tunnel-db:
	ssh -i keys/defaultec2.pem -N -L 5432:reddit-soccer-goals.cxdhgbr8e3pn.us-east-1.rds.amazonaws.com:5432 ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com -v

get-poller-logs:
	scp -i keys/defaultec2.pem ec2-user@ec2-52-7-61-91.compute-1.amazonaws.com:~/goal_poller_output.txt .

# utility commands --------------------------------------------------------------------------------------------------------
create-s3-bucket:
	aws --endpoint-url=http://localhost:4566 s3 mb s3://reddit-soccer-goals