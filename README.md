# CaTUstrophy backend [![Build Status](https://travis-ci.org/caTUstrophy/backend.svg?branch=master)](https://travis-ci.org/caTUstrophy/backend)

Backend part for our catastrophe aid tool. Written in Go.

This project provides the backend for a platform connecting people in a region suffering from a catastrophe, e.g. a natural disaster. The frontend part can be found [here](https://github.com/caTUstrophy/frontend). We develop this platform within the scope of one of our university courses, the [Programmierpraktikum: Soziale Netzwerke](https://www.cit.tu-berlin.de/menue/teaching/sommersemester_16/programmierpraktikum_soziale_netzwerke_ppsn/).

## Get it running

**1)** You have to have a working [Go installation](https://golang.org/doc/install) on your system. Preferably via your system's package manager.

**2)** Initially run
```bash
$ go get github.com/caTUstrophy/backend
```
which downloads the backend part of the project into your GOPATH.

**3)** Navigate to the project in your file system via
```bash
$ cd ${GOPATH}/src/github.com/caTUstrophy/backend
```
and execute
```bash
$ go get ./...
```
to fetch all dependencies of this project.

**4)** Create an `.env` file suited to your deployment. For this, copy the provided `.env.example` to `.env` and edit it to your needs. **Choose strong secret keys!**

**5)** Build the project via
```bash
$ go build
```

**6a)** If you are running the project the first time or after you dropped the database to start fresh, start the backend via
```bash
$ ./backend --init
```
This will create the tables and fill in some default needed content.

**6b)** Alternatively - and in the most common case - start it with
```bash
$ ./backend
```

Afterwards, the backend is reachable at `http://localhost:3001`.


## API documentation

Four roles are present in this model:
* unregistered user **(U)**: not yet present in our system
* not-logged-in user **(N)**: registered, but not authorized user
* logged-in user **(L)**: registered and authorized user
* admin **(A)**: registered, authorized and privileged user

<<<<<<< HEAD
| Functionality     			| Minimum needed privilege | HTTP verb | Endpoint       | API version | Done? |
| ------------------------- ----| ------------------------ | --------- | -------------- | ----------- | ----- |
| Registration      			| U                        | POST      | /users         | MVP         | ✔     |
| Login            				| N                        | POST      | /auth          | MVP         | ✔     |
| Renew auth token  			| L                        | GET       | /auth          | MVP         | ✔     |
| Logout            			| L                        | DELETE    | /auth          | MVP         | ✔     |
| Own profile       			| L                        | GET       | /me            | 2.0         |       |
| List offers for region x 		| A                        | GET       | /offers/x      | MVP         |       |
| List own offers   			| L                        | GET       | /me/offers     | 2.0         |       |
| List requests  for region x 	| A                        | GET       | /requests/x    | MVP         |       |
| List own requests 			| L                        | GET       | /me/requests   | 2.0         |       |
| Create offer      			| L                        | POST      | /offers        | MVP         |       |
| Create request    			| L                        | POST      | /requests      | MVP         |       |
| Update offer x    			| L                        | PUT       | /me/offers/x   | 2.0         |       |
| Update request x  			| L                        | PUT       | /me/requests/x | 2.0         |       |
| Create matching   			| A                        | POST      | /matchings     | MVP         |       |
| Get matching x    			| L                        | GET       | /matchings/x   | MVP         |       |


### Detailed request information

#### Registration

**Request:**

```
POST /users

{
    "Name": required, string
    "PreferredName": optional, string
    "Mail": required, string/email
    "Password": required, string
}
```

**Response:**

```
201 Created

{
    "ID": int
}
```

#### Login

**Request:**

```
POST /auth

{
    "Mail": required, string/email
    "Password": required, string
}
```

**Response:**

```
200 OK

{
    "AccessToken": string/jwt,
    "ExpiresIn": int
}
```

#### Renew auth token

**Request:**

```
GET /auth
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

```
200 OK

{
    "AccessToken": string/jwt,
    "ExpiresIn": int
}
```

Or - if an expired token was presented:

```
401 Unauthorized
WWW-Authenticate: Bearer realm="CaTUstrophy", error="invalid_token", error_description="<ERROR DESCRIPTION>"
```

#### Logout

**Request:**

```
DELETE /auth
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

```
200 OK
```

#### List offers

**Request:**

```
GET /offers/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**
Succes

```
200 OK
{
	"Offers": [
		{
			{
				"Name": string,
				"Tags": string array,
				"ValidityPeriod": unix timestamp,
				"Location": string,
				"User": {
					"Name": string,
					"ID": int id
				}
			}
		}, 
		...
	]
}
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


#### List requests

**Request:**

```
GET /requests/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```


**Response:**
Succes

```
200 OK
{
	"Requests": [
		{
			{
				"Name": string,
				"Tags": string array,
				"ValidityPeriod": unix timestamp,
				"Location": string,
				"User": {
					"Name": string,
					"ID": int id
				}
			}
		}, 
		...
	]
}
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```

#### Create offer

**Request:**

```
POST /offers
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
{
	"Name": required, string,
	"Tags": optional, string array,
	"ValidityPeriod": required, unix timestamp,
	"Location": required, string
}
```

***Example:***

```
POST /offers
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
{
	"Name": "hugs",
	"Tags": ["tag", "another tag"],
	"ValidityPeriod": "1464706055",
	"Location": "worldwide"
}
```

**Response:**

Success

```
200 OK
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```

***Example:***

```
400 Bad Request
{
	"Location": "User can't post for this location. (But don't expect this exact message)"
}
```

#### Create request


**Request:**

```
POST /requests
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
{
	"Name": required, string,
	"Tags": optional, string array,
	"ValidityPeriod": required, unix timestamp,
	"Location": required, string
}
```

**Response:**

Succes

```
200 OK
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```
#### Create matching

**Request:**

```
POST /matchings
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
{
	"Request": required, int id,
	"Offer": required, int id
}
```

**Response:**
Succes

```
200 OK
{
	"ID": int id
}
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```

#### Get matching x

**Request:**

```
GET /matchings/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**
Succes

```
200 OK
{
	"Request": {
		"ID": int id,
		"Name": string,
		"Tags": string array,
		"ValidityPeriod": unix timestamp,
		"Location": string
	},
	"Offer": {
		"ID": int id,
		"Name": string,
		"Tags": string array,
		"ValidityPeriod": unix timestamp,
		"Location": string
	},
}
```

Fail

```
400 Bad Request
{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```
