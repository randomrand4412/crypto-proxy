# Proxy
## Description

This is simplified version of generic proxy server. It was designed to work only with HTTP traffic. 
On start we need to define a list of edpoints that needs to be proxied, everything else will be rejected. 
Every endpoint has its own independent rate limiter, it can be easily changed to rate limiter per host
or shared between of subset of endpoints if needed.

## Architecture

### Flow:
```
                                           Proxy
 -----------------      ------------------------------------------------------------      ---------------
| External Client | -> | Request -> Middlewares -> Executor <-> goroutine -> Client | -> | Origin Server |
 -----------------      ------------------------------------------------------------      ---------------
 ```
### Request:

Request is controlled by underlying framework. I am using default golang http server. Every request is launched in its own goroutine and kept alive in memory until logic will decide to release it. 
 
### Middlewares

Http handlers that perform common logic for every request, like:
- enreach request data with redirection logic
- set auth token
- filter out unknown endpoint (not defined in config)

### Executor

In memory executor is a main component responsible for business logic. It makes calls and handles responses. 

Logic can be defined into 2 parts:
1) Fresh requests
2) Retriable requests

Every new incoming request is fresh and logic will always try to execute it immediately, actions performed for fresh requests:
- Lauch timer, used to track if request is taking too long
- Check rate limiter allowance
- Make request

If during any of phases timer is triggered, we relase incoming connection but continue running request flow. 

Requests that needs to be retried have simpler flow, we continuesly retry then until we will get some terminal state. 
They do not use timer and not bounded. 

### Callback

Proxy supports webhook on success if URI was provided on request via 'callback' query param. It will be triggered only on success. 

## Not done 

1) Deduplication of requests based on message hash
2) Cap maximum number of inflight requests in memory
3) Not all http errors are handled correctly, only subset was covered
4) Migrate from command line config to file
5) Rate limiter with sliding window
6) Persistent infligt requests
7) Rich semantics for webhooks: failure, status 
8) Droped client connections
9) Graceful shutdown
10) Maybe something else, who knows
 
## Docker run command

Build docker image first `docker build . -t proxy`.

Run image with local test endpoint
`docker run -ti -p 8080:8080 --rm proxy -t 6f6b94ab1205cf14b80eb62617d59c17 --paths "/crypto/sign,/crypto/verify" -p 8080 -o "http://host.docker.internal:9090"`

Run iamge with prod ednpoint
`docker run -ti -p 8080:8080 --rm proxy -t 6f6b94ab1205cf14b80eb62617d59c17 --paths "/crypto/sign,/crypto/verify" -p 8080 -o "https://hiring.api.synthesia.io"`