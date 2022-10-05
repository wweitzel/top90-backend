# top90-backend

Top90 is a website that populates with soccer goals in real time as they happen around the world

https://top90.io

## Applications in this repo
Goal Poller - The goal poller is a program that runs as a cron job and polls reddit.com/r/soccer for new goal videos to store in db

Server - The server is the API for the website

## Development Environment Setup
1. Install Go
2. Create a `.env` in the root directory and add the contents of `.env.sample` to it with real values
3. Start dev db and s3 in docker
```
$ docker-compose up
```
4. Install awscli
```
$ brew install awscli
```
5. Create s3 bucket in local s3. This is needed for storing goal videos locally.
```
$ aws --endpoint-url=http://localhost:4566 s3 mb s3://reddit-soccer-goals
```
6. Run db migrations
```
$ make migrate-up
```
7. Run the poller to store some goals. This will take about 2-5 min to run.
```
$ make run-poller
```
8. Go to http://localhost:8090/?pgsql=db
9. Login with the following credentials
```
username: admin
password: admin
database: redditsoccergoals
```
10. Look at the goals table in the ui and verify it has goals
11. (Optional) You can run the front end locally to see the goals https://github.com/wweitzel/top90-frontend. Note: Make sure you switch it to connect to local backend!

## Remaining Work
Finish the internal/apifootball client in order to:
- Add Premier League only capability
- Add true search capability based on team on player by matching goals to a player/team stored in db
- Show teanm schedules and rosters and click them to see the goals

## Linux Commands

#### Run docker container in background restarting automatically unless stopped
```
$ docker run -p 7171:7171 -d --restart unless-stopped top90-server-v0
```

#### Renew cert
```
$ docker-compose run --rm certbot renew
```
