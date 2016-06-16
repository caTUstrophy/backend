# CaTUstrophy backend [![Build Status](https://travis-ci.org/caTUstrophy/backend.svg?branch=master)](https://travis-ci.org/caTUstrophy/backend)

Backend part for our catastrophe aid tool. Written in Go.

This project provides the backend for a platform connecting people in a region suffering from a catastrophe, e.g. a natural disaster. The frontend part can be found [here](https://github.com/caTUstrophy/frontend). We develop this platform within the scope of one of our university courses, the [Programmierpraktikum: Soziale Netzwerke](https://www.cit.tu-berlin.de/menue/teaching/sommersemester_16/programmierpraktikum_soziale_netzwerke_ppsn/).


## Get it running

**1)** You have to have a working [Go installation](https://golang.org/doc/install) on your system. Preferably via your system's package manager.
Also, you need to have a working and running PostgreSQL instance.

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

**5)** Add PostGIS to your database. Run in your psql **(as a superuser)**
```
CREATE EXTENSION postgis;
CREATE EXTENSION postgis_topology;
```

**6)** Set up a Postgres database. Create a Postgres user and set a password for that user. Then add these information to your just created environment `.env` file. As above, have a look at the `.env.example` for a description of which values you have to set.

**7)** Build the project via
```bash
$ go build
```

**8a)** If you are running the project the first time or after you dropped the database to start fresh, start the backend via
```bash
$ ./backend --init
```
This will create the tables and fill in some default needed content.

**8b)** Alternatively - and in the most common case - start it with
```bash
$ ./backend
```

Afterwards, the backend is reachable at `http://localhost:3001`.


## Available admin user

To be able to use this backend during development we provide a default admin account. When you initially run `./backend --init`, an admin user with the following credentials will be created:

**Mail:** admin@example.org  
**Password:** CaTUstrophyAdmin123$

**Please make sure you don't create this default admin user in production!**


## API documentation

Four roles are present in this model:
* unregistered user **(U)**: not yet present in our system
* not-logged-in user **(N)**: registered, but not authorized user
* logged-in user **(L)**: registered and authorized user
* logged-in and concerned user **(C)**: user is involved in e.g. a matching
* admin **(A)**: registered, authorized and privileged user

The coloumn `Role` denotes the minimum needed privilege to use the endpoint.

| Functionality                       | Role | HTTP verb | Endpoint                     | API version | Done? |
| ----------------------------------- | ---- | --------- | ---------------------------- | ----------- | ----- |
| Login                               | N    | POST      | /auth                        | MVP         | ✔    |
| Renew auth token                    | L    | GET       | /auth                        | MVP         | ✔    |
| Logout                              | L    | DELETE    | /auth                        | MVP         | ✔    |
| Create user                         | U    | POST      | /users                       | MVP         | ✔    |
| List users                          | A    | GET       | /users                       | 3.0         |       |
| Get user `userID`                   | A    | GET       | /users/:userID               | 3.0         |       |
| Update user `userID`                | A    | PUT       | /users/:userID               | 3.0         |       |
| Create offer                        | L    | POST      | /offers                      | MVP         | ✔    |
| Get offer `offerID`                 | A    | GET       | /offers/:offerID             | 2.0         |       |
| Update offer `offerID`              | C    | PUT       | /offers/:offerID             | 3.0         |       |
| Create request                      | L    | POST      | /requests                    | MVP         | ✔    |
| Get request `requestID`             | A    | GET       | /requests/:requestID         | 2.0         |       |
| Update request `requestID`          | C    | PUT       | /requests/:requestID         | 3.0         |       |
| Create matching                     | A    | POST      | /matchings                   | MVP         | ✔    |
| Get matching `matchingID`           | C    | GET       | /matchings/:matchingID       | MVP         | ✔    |
| Update matching `matchingID`        | A    | PUT       | /matchings/:matchingID       | 3.0         |       |
| Create a region                     | L    | POST      | /regions                     | 2.0         |       |
| List regions                        | U    | GET       | /regions                     | 2.0         | ✔    |
| Get region `regionID`               | U    | GET       | /regions/:regionID           | 2.0         |       |
| Update region `regionID`            | A    | PUT       | /regions/:regionID           | 2.0         |       |
| List offers in region  `regionID`   | A    | GET       | /regions/:regionID/offers    | 2.0         |       |
| List requests in region `regionID`  | A    | GET       | /regions/:regionID/requests  | 2.0         |       |
| List matchings in region `regionID` | A    | GET       | /regions/:regionID/matchings | 2.0         |       |
| Own profile                         | L    | GET       | /me                          | 2.0         |       |
| Update own profile                  | L    | PUT       | /me                          | 2.0         |       |
| List own offers                     | L    | GET       | /me/offers                   | 2.0         |       |
| List own requests                   | L    | GET       | /me/requests                 | 2.0         |       |


### What is inside a JWT?

Inside a JWT issued by this backend, the following fields are present:

```
{
	"iss": "ralollol@bernd.orgorg",
	"iat": "2016-06-09T21:39:12+02:00",
	"nbf": "2016-06-09T21:38:12+02:00",
	"exp": "2016-06-09T22:09:12+02:00"
}
```

**iss (issuer):** Mail of user this token is issued to.  
**iat (issued at):** Time and date when token was issued. [RFC3339 date](https://www.ietf.org/rfc/rfc3339.txt)  
**nbf (not before):** Token is to be discarded when used before this time and date. [RFC3339 date](https://www.ietf.org/rfc/rfc3339.txt)  
**exp (expires):** Token is to be discarded when used after this time and date. [RFC3339 date](https://www.ietf.org/rfc/rfc3339.txt)

Please note that further identification fields may be added in the future.


### Detailed request information


#### Create user (registration)

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

Success (**Currently!** Format will change very soon!)

```
201 Created

{
	"ID": UUID v4,
	"Mail": string,
	"Name": string,
	"PreferredName": string,
	"Groups": [
		{
			"ID": UUID v4,
			"DefaultGroup": boolean,
			"Location": string,
			"Permissions": [
	            {
	                "ID": UUID v4,
	                "AccessRight": string,
	                "Description": string
	            }
            ]
		}
	]
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
    "AccessToken": string/jwt
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

{
	"ID": UUID v4
}
```


#### Own profile

**Request:**

```
GET /me
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
```

Fail

```
```


#### Update own profile

**Request:**

```
POST /me
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
}
```

**Response:**

Success

```
```

Fail

```
```


#### List offers for region x

**Request:**

```
GET /offers/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
200 OK

[
	{
		{
			"Name": string,
			"Tags": string array,
			"ValidityPeriod": RFC3339 date,
			"Location": string,
			"User": {
				"ID": UUID v4,
				"Name": string
			}
		}
	}, 
	...
]
```

Fail

```
400 Bad Request

{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


#### List own offers

**Request:**

```
GET /me/offers
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
```

Fail

```
```


#### List requests for region x

**Request:**

```
GET /requests/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
200 OK

[
	{
		{
			"Name": string,
			"Tags": string array,
			"ValidityPeriod": RFC3339 date,
			"Location": string,
			"User": {
				"ID": UUID v4,
				"Name": string
			}
		}
	}, 
	...
]
```

Fail

```
400 Bad Request

{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


#### List own requests

**Request:**

```
GET /me/requests
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
```

Fail

```
```


#### Create offer

**Request:**

```
POST /offers
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
	"Name": required, string,
	"Tags": optional, string array,
	"ValidityPeriod": required, RFC3339 date,
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
	"ValidityPeriod": "2017-11-01T22:08:41+00:00",
	"Location": "worldwide"
}
```

**Response:**

Success (**Currently!** Format will change very soon!):

```
201 Created

{
	"ID": UUID v4,
	"Location": string,
	"Name": string,
	"Tags": [
		{
			"ID": UUID v4,
			"Name": string
		},
		...
	],
	"ValidityPeriod": RFC3339 date
}
```

Fail:

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
	"ValidityPeriod": required, RFC3339 date,
	"Location": required, string
}
```

**Response:**

Success (**Currently!** Format will change very soon!):

```
201 Created

{
	"ID": UUID v4,
	"Location": string,
	"Name": string,
	"Tags": [
		{
			"ID": UUID v4,
			"Name": string
		},
		...
	],
	"ValidityPeriod": RFC3339 date
}
```

Fail:

```
400 Bad Request

{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


#### Update own offer x

**Request:**

```
PUT /me/offers/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
}
```

**Response:**

Success

```
```

Fail

```
```


#### Update own request x

**Request:**

```
PUT /me/requests/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
}
```

**Response:**

Success

```
```

Fail

```
```


#### Create matching

**Request:**

```
POST /matchings
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
	"Area": required, UUID v4,
	"Request": required, UUID v4,
	"Offer": required, UUID v4
}
```

**Response:**

Success (**Currently!** Format will change very soon!):

```
201 Created

{
	"ID": UUID v4,
	"OfferId": UUID v4,
	"RequestId": UUID v4,
	"Offer": {
		"ID": UUID v4,
		"Name": string,
		"User": {
			<currently empty user object>
		},
		"Location": string,
		"Tags": null,
		"ValidityPeriod": RFC3339 date,
		"Expired": boolean
	},
	"Request": {
		"ID": UUID v4,
		"Name": string,
		"User": {
			<currently empty user object>
		},
		"Location": string,
		"Tags": null,
		"ValidityPeriod": RFC3339 date,
		"Expired": boolean
	}
}
```

Fail:

```
400 Bad Request

{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


#### List matchings

**Request:**

```
GET /matchings
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
```

Fail

```
```


#### Get matching x

**Request:**

```
GET /matchings/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success:

```
200 OK

{
	"Request": {
		"ID": UUID v4,
		"Name": string,
		"Tags": string array,
		"ValidityPeriod": RFC3339 date,
		"Location": string
	},
	"Offer": {
		"ID": UUID v4,
		"Name": string,
		"Tags": string array,
		"ValidityPeriod": RFC3339 date,
		"Location": string
	},
}
```

Fail:

```
400 Bad Request

{
	"<FIELD NAME>": "<ERROR MESSAGE FOR THIS FIELD>"
}
```


### Create an area

**Request:**

```
POST /areas
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
}
```

**Response:**

Success

```
```

Fail

```
```


### List areas

**Request:**

```
GET /areas
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>
```

**Response:**

Success

```
```

Fail

```
```


### Update area x

```
PUT /areas/x
Authorization: Bearer <USER'S ACCESS TOKEN AS JWT>

{
}
```

**Response:**

Success

```
```

Fail

```
```
