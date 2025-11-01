#!/bin/sh

if [ "$MATRIX_HS" = '' ]; then
	echo "error: environment variable MATRIX_HS must be set" >&2
	exit 1
fi

if [ "$MATRIX_TOKEN" = '' ]; then
	echo "error: environment variable MATRIX_TOKEN must be set" >&2
	exit 2
fi

if [ "$MATRIX_ROOMID" = '' ]; then
	echo "error: environment variable MATRIX_ROOMID must be set" >&2
	exit 3
fi

if ! [ -f gh-standup ]; then
	go build -o gh-standup ./cmd/standup
fi


TXN=$(date +%s%N)

./gh-standup | \
	jq -Rs '{msgtype:"m.text", body:.}' | \
	curl -sS -X PUT "$MATRIX_HS/_matrix/client/v3/rooms/$MATRIX_ROOMID/send/m.room.message/$TXN" \
      -H "Authorization: Bearer $MATRIX_TOKEN" \
      -H "Content-Type: application/json" \
      --data @-

