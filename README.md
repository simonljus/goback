# goback
Go API-service 

uses [Gin](https://github.com/gin-gonic/gin)
## Authentication
* [gin-contrib/sessions](https://github.com/gin-contrib/sessions)

* [following auth example](https://github.com/Depado/gin-auth-example/blob/master/main.go)

## Test
* [steinfletcher/apitest](https://github.com/steinfletcher/apitest) for testing API requests and responses
# How to run
* install go
* install dependencies
* `go run api.go`
* `go test` for testing
* found at localhost:PORTNUMBER (9001 as default)
  
## API
| Description 	| Request type 	| path 	| urlparameter 	| formparameters 	| Requires authentication 	|
|-	|-	|-	|-	|-	|-	|
| PING 	| GET 	| /hellothere 	|  	|  	|  	|
| Get all messages 	| GET 	| /messages 	|  	|  	|  	|
| Create message 	| POST 	| /message 	|  	| message: the message 	| true 	|
| Edit message 	| PUT 	| /message/:messageid 	| messageid: the messageid to edit 	| message: the updated message 	| true 	|
| Delete Message 	| DELETE 	| /message/:messageid 	| messageid: the messageid to delete 	|  	| true 	|
| Sign up 	| POST 	| /signup 	|  	| username: a unique username <br>password: a password 	|  	|
| Sign in 	| POST 	| /signin 	|  	| username<br>password 	|  	|
| Sign out 	| POST 	| /signout 	|  	|  	| (true) 	|
| Get information about me 	| GET 	| /me 	|  	|  	| true 	|
| Forget me,delete user and my messages 	| DELETE 	| /me 	|  	|  	| true 	|
## TODO
* Easier install and deploy
  * Go module for installing dependencies
  * Dockerfile 
* Frontend
