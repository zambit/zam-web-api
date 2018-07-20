#/bin/bash

echo 'env: staging
db:
    uri: '$STAGING_DB_URI'
server:
    storage:
        uri: '$STAGING_REDIS_URI'
    auth:
        tokenstorage: jwtpersistent
    generator:
        codealphabet: 4
    jwt:
        secret: '$STAGING_SECRET'
        method: HS256'