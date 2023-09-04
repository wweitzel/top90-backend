# top90-backend

top90 is a website that populates with soccer goals in real time as they happen around the world.

https://top90.io

# Contributing Guide
Anyone is welcome to submit a PR. PRs should be tested and verified locally first if possible.

## Applications in this repo
- server - The API for the website.
- goal_poller - Polls reddit.com/r/soccer for new goals/videos to store in db/s3.
- apifootball_ingest - Gets up to date team, league, and fixture data.

## Running Locally
1. Install Go and Docker if you do not have them installed.
```
$ brew install go
$ brew install --cask docker
```
2. Install awscli, golang-migrate, and ffmpeg.
```
$ brew install awscli
$ brew install golang-migrate
$ brew install ffmpeg
```
3. Create local environment files.
```
cp .env.sample .env
cp .env.docker.sample .env.docker
```
4. Start dev db and s3 in docker.
```
$ docker-compose up
```
5. Seed local database.
```
$ make seed
```
6. Run the server
```
$ go run ./cmd/server/...
```
7. Go to http://127.0.0.1:7171/goals in a browser to verify backend is running and returning data.

## Viewing local database
Part of the docker compose runs a database viewer. Go to http://localhost:8090/?pgsql=db logging in with the following to see it.
```
username: admin
password: admin
database: redditsoccergoals
```

## Tests
```
# Make sure docker daemon is running
$ go test ./...
```

## Debugging
For vscode, make a `.vscode/launch.json` file and paste the following in it.
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "./cmd/server",
            "cwd": "./"
        }
    ]
}
```
The above configuration will give you the options to debug in the "Run and Debug" tab of vscode.

## Creating a New Migration
Run the following and modify the generated files. See https://github.com/golang-migrate/migrate for details.
```
migrate create -ext sql -dir internal/db/migrations -seq name_of_migration
```

## Resetting Data
If you want to start fresh, you can easily wipe all your data by deleting the two folders in `docker-data` or simply run `make seed` which will repopulate the local db/s3. 

## Renew cert
Command to renew cert on the ec2
```
$ docker-compose run --rm certbot renew && docker-compose restart
```
