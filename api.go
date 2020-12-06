package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func (user *User) toDTO() UserDTO {
	return UserDTO{Id: user.Id, Username: user.Username}
}

func (message *Message) toDTO(userId int64) MessageDTO {
	var showId int64
	if message.userId == userId {
		showId = userId
	}
	return MessageDTO{Id: message.Id, Text: message.Text, UserId: showId}
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
	m := Message{Id: seq, Text: text, userId: userId}
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
func deleteMessageWithId(messageId int64) {
	delete(_messages, messageId)
}
func deleteMessage(userId, messageId int64) error {
	message, err := getMessageById(messageId)
	if err != nil {
		return nil
	} else if message.userId != userId {
		return fmt.Errorf("Not authorized to delte message")
	} else {
		deleteMessageWithId(messageId)
		return nil
	}

}
func editMessage(dto MessageDTO) (MessageDTO, error) {
	if message, err := getUserMessage(dto.UserId, dto.Id); err != nil {
		return MessageDTO{}, err
	} else {
		message.Text = dto.Text
		_messages[dto.Id] = message
		return dto, nil
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
		fmt.Println("userkey", userId)
		return -1, fmt.Errorf("Unathorized")
	} else {
		fmt.Println("useridnil", userId)
		return userId.(int64), nil
	}
}
func authHandler(c *gin.Context) {
	fmt.Println("authhandler")
	if _, err := getUserIdFromSession(c); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	} else {
		c.Next()
	}

}
func signout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.JSON(http.StatusNoContent, gin.H{"message": "Signed out"})
}
func signin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserIdFromSession(c); err == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Sign out first"})
	} else if user, err := getUserByUsername(username); err == nil && password == user.passsword {
		session := sessions.Default(c)
		fmt.Println("signed in ", user)
		session.Set(userkey, user.Id)
		sessionError := session.Save()
		if sessionError != nil {
			fmt.Println(sessionError.Error())
		}

		c.JSON(http.StatusOK, gin.H{"message": "signed in"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The username and password did not match"})
	}

}
func createUser(username string, password string) User {
	seq := getSequenceNumber()
	user := User{seq, username, password}
	_users[seq] = user
	_usernames[username] = seq
	return user
}
func deleteUser(userId int64) {
	if user, err := getUserById(userId); err == nil {
		delete(_users, userId)
		delete(_usernames, user.Username)
	}
}
func signup(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if _, err := getUserIdFromSession(c); err == nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Sign out first"})
	} else if _, err := getUserByUsername(username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
	} else {
		user := createUser(username, password)
		c.JSON(http.StatusOK, gin.H{"message": "user created", "user": user})
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
func getMe(c *gin.Context) {
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
func deleteMe(c *gin.Context) {
	// already checked by middleware
	userId, _ := getUserIdFromSession(c)
	session := sessions.Default(c)
	session.Clear()
	deleteMessagesByUser(userId)
	deleteUser(userId)
	c.JSON(http.StatusNoContent, gin.H{"message": "Deleted"})
}
func createEngine() *gin.Engine {
	r := gin.New()
	store := cookie.NewStore([]byte(fmt.Sprintf("secret%d", time.Now().Unix())))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/hellothere", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "General Kenobi"})
	})
	r.POST("/signin", signin)
	r.POST("/signup", signup)
	r.POST("/signout", signout)

	r.GET("/messages", func(c *gin.Context) {
		var userId = int64(0)
		sessionUserId, err := getUserIdFromSession(c)
		if err == nil {
			userId = sessionUserId
		}
		messages := getMessages(userId)
		c.JSON(http.StatusOK, messages)
	})
	authorizedMe := r.Group("/me/")
	authorizedMe.Use(authHandler)
	authorizedMe.DELETE("", deleteMe)
	authorizedMe.GET("", getMe)
	authorizedMessage := r.Group("/message")
	authorizedMessage.Use(authHandler)
	authorizedMessage.POST("/:message", func(c *gin.Context) {
		messageText := c.Param("message")
		// already caught by middleware
		userId, _ := getUserIdFromSession(c)
		message := createMessage(userId, messageText)
		c.JSON(http.StatusCreated, message)
	})
	authorizedMessage.PUT("/:messageid/:message", func(c *gin.Context) {
		userId, _ := getUserIdFromSession(c)
		if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
		} else if text := c.Param("message"); len(text) != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message is required"})
		} else if message, messageError := editMessage(MessageDTO{id, text, userId}); messageError != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit message"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": message})
		}
	})
	authorizedMessage.DELETE("/:messageid/", func(c *gin.Context) {
		userId, _ := getUserIdFromSession(c)
		if id, err := strconv.ParseInt(c.Param("messageid"), 10, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid message Id is required"})
		} else if err := deleteMessage(userId, id); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete message"})
		} else {
			c.JSON(http.StatusNoContent, gin.H{"message": "Message does not exist"})
		}
	})
	return r
}
func main() {
	r := createEngine()
	r.Run()
}
