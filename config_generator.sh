#/bin/bash

echo 'env: staging
db:
    uri: '$STAGING_DB_URI'
server:
    storage:
        uri: '$STAGING_REDIS_URI'
    auth:
        tokenstorage: jwtpersistent
    jwt:
        secret: '$STAGING_SECRET'
        method: HS256
    notificator:
        url: '$NOTIFICATIONS_URL'

isc:
    brokeruri: '$BROKER_URI'
    statsenabled: true
    statspath: /internal/stats
'