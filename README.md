# goback
Go API-service 

uses [Gin](https://github.com/gin-gonic/gin)
## Authentication
* [gin-contrib/sessions](https://github.com/gin-contrib/sessions)

* [following auth example](https://github.com/Depado/gin-auth-example/blob/master/main.go)

# How to run
* install go
* install dependencies
* `go run api.go`
* `go test` for testing
  
## Features
* Get all messages
* Add message
* Edit message
* Delete message
* Authentication
  * signup
  * signin
  * signout
  * Delete user and all their messages
## TODO
* Test cases
  * API
* Easier install and deploy
  * Go module for installing dependencies
  * Dockerfile 
* Frontend
