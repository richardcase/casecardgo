# CaseCard In Go [![Build Status](https://travis-ci.org/richardcase/casecardgo.svg?branch=master)](https://travis-ci.org/richardcase/casecardgo)

A sample prepaid card written in Go. Its very rough and ready!!!

## To start locally

User docker compsose to start the services:
```bash
docker-compose up
```

Then use the postman collection file in the artefacts folder.

## NATS Endpoints
- http://localhost:8222/varz
- http://localhost:8222/connz
- http://localhost:8222/subscriptionsz
- http://localhost:8222/routez

## Scratch

```bash
docker run -d -p 27017:27017 mongo
docker run -d -p 4222:4222 -p 6222:6222 -p 8222:8222 --name nats-main nats
```
