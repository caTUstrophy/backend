#!/bin/bash
#Test cases:

echo ''
echo '### TEST: Offer ###'

# -> create user
echo 'POST /users'
curl -s -H "Content-Type: application/json" -X POST -d '{"Name": "jery", "PreferredName": "jery", "Mail": "jery@gmx.de", "Password": "yoloyoloyoloyolo!12"}' http://localhost:3001/users

# -> login
echo 'POST /auth'
TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"Mail":"jery@gmx.de", "Password":"yoloyoloyoloyolo!12"}' http://localhost:3001/auth)
TOKEN=$(echo $TOKEN | cut -d'"' -f 4)
#echo "$TOKEN"
echo ''


# -> get request
echo 'GET /requests/Berlin'
REQ=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET  http://localhost:3001/requests/Berlin)
echo "$REQ"
echo 'GET /requests/DoesNotExist'
REQ=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET  http://localhost:3001/requests/DoesNotExist)
echo "$REQ"
echo ''

# -> get offers
echo 'GET /offers/Berlin'
OFFER=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET  http://localhost:3001/offers/Berlin)
echo "$OFFER"
echo 'GET /offers/DoesNotExist'
OFFER=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X GET  http://localhost:3001/offers/DoesNotExist)
echo "$OFFER"
