#!/bin/bash
#Test cases:

echo ''
echo '### TEST: Offer ###'

## --------- LOGIN ----------
# -> create user
echo 'POST /users'
curl -s -H "Content-Type: application/json" -X POST -d '{"Name": "jery", "PreferredName": "jery", "Mail": "jery@gmx.de", "Password": "yoloyoloyoloyolo!12"}' http://localhost:3001/users

# -> login
echo 'POST /auth'
TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"Mail":"jery@gmx.de", "Password":"yoloyoloyoloyolo!12"}' http://localhost:3001/auth)
TOKEN=$(echo $TOKEN | cut -d'"' -f 4)
#echo "$TOKEN"
echo ''


## -------- TEST : LIST

# -> list offer in region
echo 'GET /requests/Berlin'
REQ=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET http://localhost:3001/requests/Berlin)
echo "$REQ"

# -> list offer in region
echo 'GET /requests/DoesNotExist'
REQ=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET http://localhost:3001/requests/DoesNotExist)
echo "$REQ"
echo ''

# -> list offer in region
echo 'GET /offers/Berlin'
OFFER=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET http://localhost:3001/offers/Berlin)
echo "$OFFER"

# -> list offer in region
echo 'GET /offers/DoesNotExist'
OFFER=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET http://localhost:3001/offers/DoesNotExist)
echo "$OFFER"
