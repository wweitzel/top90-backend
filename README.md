# top90-backend

top90 is a website that populates with soccer goals in real time as they happen around the world.

https://top90.io

## Applications in this repo
- server - The server is the API for the website

- goal_poller - The goal poller is a script that runs as a cron job and polls reddit.com/r/soccer for new goals/videos to store in db/s3

- apifootball_ingest - Script to store apifootball data in the database

# Contributing Guide
Anyone is welcome to submit a PR. PRs should be tested and verified locally first if possible.

## Running Locally
1. Install dependencies. If using homebrew you can run the following.
```
$ brew install go
$ brew install --cask docker
$ brew install awscli
$ brew install golang-migrate
$ brew install ffmpeg
```
2. Create a `.env` in the root directory and add the contents of `.env.sample` to it.
```
cp .env.sample .env
```
3. Start dev db and s3 in docker.
```
# Make sure docker daemon is running
$ docker-compose up
```
4. Seed local database
```
$ make seed
# Answer "y" to the prompt when you see it
```
5. Go to http://127.0.0.1:7171/goals in a browser to verify backend is running and returning data.
6. Thats it. If interested, the frontend repo is here https://github.com/wweitzel/top90-frontend.

## Viewing local database
Part of the docker compose runs a database viewer. Go to http://localhost:8090/?pgsql=db logging in with the following to see it.
```
username: admin
password: admin
database: redditsoccergoals
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
