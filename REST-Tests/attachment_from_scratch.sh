#!/usr/bin/env bash

curl -X POST http://localhost:3456/api/v1/register -H 'Content-Type: application/json' -d '{"username":"demo","password":"demo","email":"demo@vikunja.io"}'
BEARER=`curl -X POST -H 'Content-Type: application/json' -d '{"username": "demo", "password":"demo"}' localhost:3456/api/v1/login | jq -r '.token'`

echo "Bearer: $BEARER"

curl -X POST localhost:3456/api/v1/tokenTest -H "Authorization: Bearer $BEARER"

curl -X PUT localhost:3456/api/v1/namespaces/1/lists  -H 'Content-Type: application/json' -H "Authorization: Bearer $BEARER" -d '{"title":"lorem"}'
curl -X PUT localhost:3456/api/v1/lists/1  -H 'Content-Type: application/json' -H "Authorization: Bearer $BEARER" -d '{"text":"lorem"}'
curl -X PUT -H "Authorization: Bearer $BEARER" localhost:3456/api/v1/tasks/1/attachments -F 'files=@/home/konrad/Pictures/Wallpaper/greg-rakozy-_Q4mepyyjMw-unsplash.jpg'
