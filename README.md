# top90-backend

Top90 is a website that populates with soccer goals in real time as they happen around the world

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
```
2. Create a `.env` in the root directory and add the contents of `.env.sample` to it with real values
```
cp .env.sample .env
```
3. Start dev db and s3 in docker.
```
# Make sure docker daemon is running
$ docker-compose up
```
4. Seed local data

NOTE: You need to get an api key from https://rapidapi.com/api-sports/api/api-football/ and set it as API_FOOTBALL_RAPID_API_KEY in your .env for the seed to work!
```
$ make seed
# Note: Answer "y" to the promopt when you see it
```
5. Go to http://localhost:8090/?pgsql=db
6. Login with the following credentials
```
username: admin
password: admin
database: redditsoccergoals
```
7. Look at the tables in the UI and verify they have data
8. (Optional) You can run the front end locally to see the goals https://github.com/wweitzel/top90-frontend. Note: Make sure you switch it to connect to local backend!

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
        },
        {
            "name": "Debug Poller",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "./cmd/poller",
            "cwd": "./"
        },
        {
            "name": "ApiFootball Ingest",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "./cmd/apifootball_ingest",
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
If you want to start fresh, you can easily wipe all your data by deleting the two folders in `docker-data` or simpley run `make seed` which will repopulate the local db/s3. 

# Remaining Work
Finish the internal/apifootball client in order to:
- Add Premier League only capability
- Add true search capability based on team on player by matching goals to a player/team stored in db
- Show team schedules and rosters and click them to see the goals
- Use int64 for all id columns
- Convert createdAt fields to timestamp with timezone

## Leagues Supported
- England - Premier League
- Italy - Serie A
- Spain - La Liga
- Germany - Bundesliga
- France - Ligue 1
- World - UEFA Champions League
- World - UEFA Europa League
- World - World Cup

## Linux Commands

#### Run docker container in background restarting automatically unless stopped
```
$ docker run -p 7171:7171 -d --restart unless-stopped top90-server-v0
```

#### Renew cert
```
$ docker-compose run --rm certbot renew && docker-compose restart
```
