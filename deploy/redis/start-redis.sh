#!/bin/sh
echo "Starting primary in ${FLY_REGION}"
redis-server /usr/local/etc/redis/redis.conf \
	--requirepass "${REDIS_PASSWORD}"
