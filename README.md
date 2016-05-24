[![Build Status](https://travis-ci.org/caTUstrophy/backend.svg?branch=master)](https://travis-ci.org/caTUstrophy/backend)

# CaTUstrophy backend

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

**4)** Build the project via
```bash
$ go build
```

**5)** And finally start it with
```bash
$ ./backend
```

## API documentation

Three roles are present in this model:
* unregistered user (U): not yet present in our system
* not-logged-in user (N): registered, but not authorized user
* logged-in user (L): registered and successfully authorized user
* admin (A): registered, successfully authorized and privileged user

| Functionality | Needed privilege | HTTP verb | Endpoint | API version |
| ------------- | ---------------- | --------- | -------- | ----------- |
| Registration  | U                | POST      | /users   | MVP         |
| Login         | N                | POST      | /auth    | MVP         |
| Renew auth token | L             | GET       | /auth    | MVP         |
| Logout        | L                | DELETE    | /auth    | MVP         |
| Own profile   | L                | GET       | /me      | 2.0         |
| List offers   | A                | GET       | /offers  |  MVP        |
| List own offers | L              | GET       | /me/offers |  2.0      |
| List requests | A                | GET       | /requests  | MVP       |
| List own requests | L            | GET       | /me/requests | 2.0     |
| Create offer  | L                | POST      | /offers  | MVP         |
| Create request | L               | POST      | /requests  | MVP       |
| Update offer x | L               | PUT       | /me/offers/x | 2.0     |
| Update request x | L             | PUT       | /me/requests/x | 2.0   |
| Create matching | A              | POST      | /matchings | MVP       |
| Get matching x | L               | GET       | /matchings/x | MVP     |

### Detailed request information