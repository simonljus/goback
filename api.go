package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const userkey = "user"

type User struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	passsword string
}
type UserDTO struct {
	id       int64
	username string
}
type Message struct {
	id     int64
	text   string
	userId int64
}
type MessageDTO struct {
	Id     int64  `json:"id" form:"id" binding:"required"`
	Text   string `json:"text" form:"text" binding:"required"`
	UserId int64  `json:"userId" form:"userId" binding:"required"`
}

func (user *User) toDTO() UserDTO {
	return UserDTO{id: user.Id, username: user.Username}
}

func (message *Message) toDTO(userId int64) MessageDTO {
	var showId int64
	if message.userId == userId {
		showId = userId
	}
	return MessageDTO{Id: message.id, Text: message.text, UserId: showId}
}

const PUBLICUSER = int64(0)

var _seq = int64(0)
var _usernames = make(map[string]int64)
var _users = make(map[int64]User)
var _messages = make(map[int64]Message)

func getSequenceNumber() int64 {
	_seq = _seq + 1
	return _seq
}
func getUserByUsername(username string) (User, error) {
	if userId, exists := _usernames[username]; !exists {
		return User{}, fmt.Errorf("User does not exist")
	} else {
		return getUserById(userId)
	}
}
func getUserById(userId int64) (User, error) {
	if user, exists := _users[userId]; !exists {
		return user, fmt.Errorf("User does not exist")
	} else {
		return user, nil
	}
}
func createMessage(userId int64, text string) MessageDTO {
	seq := getSequenceNumber()
	m := Message{id: seq, text: text, userId: userId}
	_messages[seq] = m
	return m.toDTO(userId)
}
func getMessages(userId int64) []MessageDTO {
	mArray := make([]MessageDTO, 0, len(_messages))
	for _, message := range _messages {
		mArray = append(mArray, message.toDTO(userId))
	}
	return mArray
}
func getMessageById(messageId int64) (Message, error) {
	if message, exists := _messages[messageId]; !exists {
		return message, fmt.Errorf("Message does not exist")
	} else {
		return message, nil
	}
}
func getUserMessage(userId, messageId int64) (Message, error) {
	if message, err := getMessageById(messageId); err != nil {
		if message.userId == userId {
			return message, nil
		}
	}
	return Message{}, errors.New(fmt.Sprintf("message %d does not exist", messageId))
}
func deleteMessage(userId, messageId int64) error {
	message, err := getMessageById(messageId)
	if err != nil {
		return nil
	} else if message.userId != userId {
		return fmt.Errorf("Not authorized to delte message")
	} else {
		delete(_messages, messageId)
		return nil
	}

}
func editMessage(dto MessageDTO) (MessageDTO, error) {
	if message, err := getUserMessage(dto.UserId, dto.Id); err != nil {
		return MessageDTO{}, err
	} else {
		message.text = dto.Text
		_messages[dto.Id] = message
		return dto, nil
	}
}
func getUserFromSession(c *gin.Context) (UserDTO, error) {
	session := sessions.Default(c)
	if user := session.Get(userkey); user == nil {
		return UserDTO{}, fmt.Errorf("Unathorized")
	} else {
		return user.(UserDTO), nil
	}
}
func authHandler(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	fmt.Println("executing", user)
	if user == nil {
		// Abort the request with the appropriate error code
		fmt.Print("Not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	c.Next()
}
func signout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out"})
}
func signin(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserFromSession(c); err == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Sign out first"})
	} else if user, err := getUserByUsername(username); err == nil && password == user.passsword {
		session.Set(userkey, user.toDTO())
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "signed in"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
	}

}
func createUser(username string, password string) User {
	seq := getSequenceNumber()
	user := User{seq, username, password}
	_users[seq] = user
	_usernames[username] = seq
	return user
}
func signup(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserFromSession(c); err == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Sign out first"})
	} else if _, err := getUserByUsername(username); err == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "User already exists"})
	} else {
		user := createUser(username, password)
		c.JSON(http.StatusOK, gin.H{"message": "user created", "user": user})
	}

}
func createEngine() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	authorized := r.Group("/message")
	authorized.Use(authHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.POST("/signin", signin)
	r.POST("/signup", signup)
	r.POST("/signout", signout)
	r.GET("/messages", func(c *gin.Context) {
		var userId = int64(0)
		userDTO, err := getUserFromSession(c)
		if err == nil {
			userId = userDTO.id
		}
		messages := getMessages(userId)
		c.JSON(200, messages)
	})
	authorized.POST("/:message", func(c *gin.Context) {
		messageText := c.Param("message")
		// already caught by middleware
		userDTO, _ := getUserFromSession(c)
		message := createMessage(userDTO.id, messageText)
		c.JSON(200, message)
	})
	authorized.PUT("/:messageid/:message", func(c *gin.Context) {
		userDTO, _ := getUserFromSession(c)
		if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
		} else if text := c.Param("message"); len(text) != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message is required"})
		} else if message, messageError := editMessage(MessageDTO{id, text, userDTO.id}); messageError != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Cannot edit message"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": message})
		}
	})
	authorized.DELETE("/:messageid/", func(c *gin.Context) {
		userDTO, _ := getUserFromSession(c)
		if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
		} else if err := deleteMessage(userDTO.id, id); err != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Cannot delete message"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Message does not exist"})
		}
	})
	return r
}
func main() {
	r := createEngine()
	r.Run()
}
