# game-library-auth
Is an authentication service for game-library app

### Usage with `Make`:
    build           builds app
    run             runs app
    test            runs tests for the whole project
    lint            runs golangci-lint

    dockerrunpg     runs postgres server with 'auth' db in docker container
    migrate         applies all migrations to database
    rollback        roll backs one last migration of database
    seed            applies seed data (roles, admin user) to database

    keygen          creates private/public key pair files

    dockerbuildauth builds auth app docker image
    dockerrunauth   runs auth app in docker container
    dockerbuildmng  builds manage app docker image
    dockerrunmng-m  applies migrations to database using docker manage image
    dockerrunmng-r  rollbacks one last migration using docker manage image
    dockerrunmng-s  seeds test data to database using docker manage image

### Endpoints:
    /signup         [POST]  - creates new user
    /signin         [POST]  - checks user credentials and returns access token
    /token/verify   [POST]  - checks validity of provided JWT
    
    /readiness      [GET]   - checks if app is ready
    /liveness       [GET]   - checks if app is up
