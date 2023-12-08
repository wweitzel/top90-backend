# top90-backend

top90 is a website that populates with soccer goals in real time as they happen around the world.

https://top90.io

# Contributing Guide
Anyone is welcome to submit a PR. PRs should be tested and verified locally first if possible.

## Applications in this repo
- api - The API for the website.
- scraper - Scrapes reddit.com/r/soccer for new videos. Stores them in a database.
- syncdata - Gets the latest league, team, and fixture data from apifootball.

## Running Locally
1. Install Go and Docker if you do not have them installed.
2. Run the following command to create local environment files.
```
cp .env.sample .env && cp .env.docker.sample .env.docker
```
3. Start dev db and s3 in docker.
```
docker-compose up
```
4. In a new terminal, seed local database
```
make seed
```
5. Run the api
```
go run ./cmd/api
```

## Tests
Make sure docker is running. The tests spin up pg instances for integration testing.
```
go test ./...
```

## Viewing local database
Part of the docker compose runs a database viewer. Go to http://localhost:8090/?pgsql=db logging in with the following to see it.
```
username: admin
password: admin
database: redditsoccergoals
```

## Creating a New Migration
Run the following and modify the generated files. See https://github.com/golang-migrate/migrate for details.
```
brew install golang-migrate
migrate create -ext sql -dir internal/db/migrations -seq name_of_migration
```
