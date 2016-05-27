#!/bin/bash
#Test cases:

echo "Tests for user registration:"
read -p "Enter an email adress that is not registered in the system:" email
## Correct request
echo "This is the only request that shall work: "
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "'$email'", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users
read -p "Enter an email adress that is not registered in the system:" email

## No valid Email
echo "Try to register with a non valid email: "

echo "Not containing an '@'"
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "niels.warnckegmail.com", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users

echo "Not containing a host like <string>.<string>"
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "niels.warncke@gmailcom", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users

echo "Not containing a string before the '@'"
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "@gmail.com", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users

## False JSON attributes
echo "Missing JSON attribut 'Name'"
curl -H "Content-Type: application/json" -X POST -d '{"PreferredName": "niels", "Mail": "'$email'", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users

echo "Containing unexpected JSOn attribut 'foo'"
curl -H "Content-Type: application/json" -X POST -d '{"foo": "bar", "Name": "niels", "PreferredName": "niels", "Mail": "'$email'", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users

## Not secure password
echo "Unsecure password: "

echo "Password: test1! (too short)"
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "'$email'", "Password": "test1!"}' http://localhost:3001/users

echo "Password: testtesttesttesttesttest (only letters)"
curl -H "Content-Type: application/json" -X POST -d '{"Name": "niels", "PreferredName": "niels", "Mail": "'$email'", "Password": "testtesttesttesttesttest"}' http://localhost:3001/users

## Preferred Name including SQL Request
echo "Preferred Name including SQL Request: "
curl -H "Content-Type: application/json" -X POST -d "{\"Name\": \"niels\", \"PreferredName' AND FALSE ": "niels", "Mail": "'$email'", "Password": "test123dijaOIUWEFO:;:;:;U45566"}' http://localhost:3001/users


