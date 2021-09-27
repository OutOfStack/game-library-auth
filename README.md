# game-library-auth
Is an authentication service for game-library app

### Usage:
    make build          builds app
    make run            runs app
    make dockerrunpg    runs postgres server with 'auth' db in docker container
    make migrate        applies all migrations to database
    make rollback       roll backs one last migration of database
    make seed           applies seed data (roles, admin user) to database
