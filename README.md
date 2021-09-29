# game-library-auth
Is an authentication service for game-library app

### Usage with `Make`:
    build          builds app
    run            runs app
    dockerrunpg    runs postgres server with 'auth' db in docker container
    migrate        applies all migrations to database
    rollback       roll backs one last migration of database
    seed           applies seed data (roles, admin user) to database
    keygen         creates private/public key pair files
