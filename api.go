package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const PORTNUMBER = 9001

type User struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	passsword string
}
type UserDTO struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
}
type Message struct {
	Id     int64  `json:"id"`
	Text   string `json:"text"`
	userId int64
}
type MessageDTO struct {
	Id     int64  `json:"id" form:"id" binding:"required"`
	Text   string `json:"text" form:"text" binding:"required"`
	UserId int64  `json:"userId" form:"userId" binding:"required"`
}

const PUBLICUSER = int64(0)
const userkey = "user"

var _seq = int64(0)
var _usernames = make(map[string]int64)
var _users = make(map[int64]User)
var _messages = make(map[int64]Message)

func reset() {
	_seq = int64(0)
	_usernames = make(map[string]int64)
	_users = make(map[int64]User)
	_messages = make(map[int64]Message)
}

func (user *User) toDTO() UserDTO {
	return UserDTO{Id: user.Id, Username: user.Username}
}

func (message *Message) toDTO(userId int64) MessageDTO {
	showId := int64(-1)
	if message.userId == userId {
		showId = userId
	}
	return MessageDTO{Id: message.Id, Text: message.Text, UserId: showId}
}

func getSequenceNumber() int64 {
	_seq += 1
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
func createMessage(userId int64, text string) Message {
	seq := getSequenceNumber()
	m := Message{Id: seq, Text: text, userId: userId}
	_messages[seq] = m
	return m
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
		return Message{}, err
	} else if message.userId != userId {
		return Message{}, fmt.Errorf("Message is made by another user")
	} else {
		return message, nil
	}
}
func deleteMessageWithId(messageId int64) {
	delete(_messages, messageId)
}
func deleteMessage(userId, messageId int64) error {
	if message, err := getMessageById(messageId); err != nil {
		return nil
	} else if message.userId != userId {
		return fmt.Errorf("Not authorized to delte message")
	} else {
		deleteMessageWithId(messageId)
		return nil
	}

}
func updateMessage(dto MessageDTO) (Message, error) {
	if message, err := getUserMessage(dto.UserId, dto.Id); err != nil {
		return Message{}, err
	} else {
		message.Text = dto.Text
		_messages[message.Id] = message
		return message, nil
	}
}

func createUser(username string, password string) (User, error) {
	if len(username) == 0 || len(password) == 0 {
		return User{}, fmt.Errorf("Username and password are required")
	} else if _, exists := _usernames[username]; exists {
		return User{}, fmt.Errorf("User already exists")
	}
	seq := getSequenceNumber()
	user := User{seq, username, password}
	_users[seq] = user
	_usernames[username] = seq
	return user, nil

}
func deleteUser(userId int64) {
	if user, err := getUserById(userId); err == nil {
		delete(_users, userId)
		delete(_usernames, user.Username)
	}
}

func getMessagesByUser(userId int64) []Message {
	messages := []Message{}

	return messages
}
func deleteMessagesByUser(userId int64) {
	for messageId, message := range _messages {
		if message.userId == userId {
			deleteMessageWithId(messageId)
		}
	}
}

func getUserFromSession(c *gin.Context) (User, error) {
	if userId, err := getUserIdFromSession(c); err != nil {
		return User{}, err
	} else if user, err := getUserById(userId); err != nil {
		return User{}, err
	} else {
		return user, nil
	}
}
func getUserIdFromSession(c *gin.Context) (int64, error) {
	session := sessions.Default(c)
	if userId := session.Get(userkey); userId == nil {
		return -1, fmt.Errorf("Unathorized")
	} else {
		return userId.(int64), nil
	}
}
func authHandler(c *gin.Context) {
	if _, err := getUserIdFromSession(c); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	} else {
		c.Next()
	}

}
func signoutReq(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.JSON(http.StatusNoContent, gin.H{"message": "Signed out"})
}
func signinReq(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserIdFromSession(c); err == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Sign out first"})
	} else if user, err := getUserByUsername(username); err == nil && password == user.passsword {
		session := sessions.Default(c)
		session.Set(userkey, user.Id)
		sessionError := session.Save()
		if sessionError != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "signed in"})
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The username and password did not match"})
	}

}

func signupReq(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserIdFromSession(c); err == nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Sign out first"})
	} else if len(username) == 0 {
		c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Username cannot be empty"})
	} else if len(password) == 0 {
		c.JSON(http.StatusPreconditionRequired, gin.H{"error": "Password cannot be empty"})
	} else if user, err := createUser(username, password); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot create user"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "user created", "user": user})
	}

}

func getUserReq(c *gin.Context) {
	user, err := getUserFromSession(c)
	if err != nil {
		// can happen if multiple users uses the same user and one removes it
		session := sessions.Default(c)
		session.Clear()
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": user, "messages": getMessagesByUser(user.Id)})
	}

}
func updateMessageReq(c *gin.Context) {
	userId, _ := getUserIdFromSession(c)
	if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
	} else if text := c.PostForm("message"); len(text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message text is required"})
	} else if message, messageError := updateMessage(MessageDTO{id, text, userId}); messageError != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit message"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message.toDTO(userId)})
	}
}
func deleteUserReq(c *gin.Context) {
	// already checked by middleware
	userId, _ := getUserIdFromSession(c)
	session := sessions.Default(c)
	session.Clear()
	deleteMessagesByUser(userId)
	deleteUser(userId)
	c.JSON(http.StatusNoContent, gin.H{"message": "Deleted"})
}
func getMessagesReq(c *gin.Context) {
	var userId = int64(0)
	sessionUserId, err := getUserIdFromSession(c)
	if err == nil {
		userId = sessionUserId
	}
	messages := getMessages(userId)
	c.JSON(http.StatusOK, messages)
}
func createMessageReq(c *gin.Context) {
	messageText := c.PostForm("message")
	if len(messageText) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message is required"})
		return
	}

	// already caught by middleware
	userId, _ := getUserIdFromSession(c)
	message := createMessage(userId, messageText)
	c.JSON(http.StatusCreated, message.toDTO(userId))
}
func deleteMessageReq(c *gin.Context) {
	userId, _ := getUserIdFromSession(c)
	if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
	} else if err := deleteMessage(userId, id); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete message"})
	} else {
		c.JSON(http.StatusNoContent, gin.H{"message": "Message does not exist"})
	}
}
func pingReq(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "General Kenobi"})
}
func getSecret() string {
	return fmt.Sprintf("secret%d", time.Now().Unix())
}
func createEngine() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte(getSecret()))
	store.Options(sessions.Options{HttpOnly: true})
	r.Use(sessions.Sessions("mysession", store))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/hellothere", pingReq)
	r.POST("/signin", signinReq)
	r.POST("/signup", signupReq)
	r.POST("/signout", signoutReq)

	r.GET("/messages", getMessagesReq)
	authorizedMe := r.Group("/me")
	authorizedMe.Use(authHandler)
	authorizedMe.DELETE("", deleteUserReq)
	authorizedMe.GET("", getUserReq)
	authorizedMessage := r.Group("/message")
	authorizedMessage.Use(authHandler)
	authorizedMessage.POST("", createMessageReq)
	authorizedMessage.PUT("/:messageid", updateMessageReq)
	authorizedMessage.DELETE("/:messageid", deleteMessageReq)
	return r
}
func StartServer() {
	r := createEngine()
	r.Run(fmt.Sprintf(":%d", PORTNUMBER))
}
func main() {
	StartServer()
}
