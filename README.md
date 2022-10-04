# top90-backend

Top90 is a website that populates with soccer goals in real time as they happen around the world

https://top90.io

## Applications in this repo
1. Goal Poller - The goal poller is a program that runs as a cron job and polls reddit.com/r/soccer for new goal videos to store in db
2. Server - The server is the API for the website

## Development Environment Setup
1. Install Go
2. Create a `.env` in the root directory and add the contents of `.env.sample` to it with real values
3. Create `/keys` in the root directory and add the `defaultec2.pem` cert for DB tunneling

## Linux Commands

#### Run docker container in background restarting automatically unless stopped
```
$ docker run -p 7171:7171 -d --restart unless-stopped top90-server-v0
```

#### Renew cert
```
$ docker-compose run --rm certbot renew
```

## Remaining Work
- Finish the internal/apifootball client in order to:
    - Add Premier League only capability
    - Add true search capability based on team on player by matching goals to a player/team stored in db
    - Show teanm schedules and rosters and click them to see the goals
