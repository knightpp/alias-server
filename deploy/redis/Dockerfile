FROM docker.io/redis:7-alpine

RUN apk add --no-cache bind-tools

COPY redis.conf /usr/local/etc/redis/redis.conf
COPY start-redis.sh /usr/bin/

RUN chmod +x /usr/bin/start-redis.sh

CMD ["start-redis.sh"]