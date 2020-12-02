package main

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type User struct {
	id   int64
	name string
}
type Message struct {
	id   int64
	text string
	user *User
}
type MessageDTO struct {
	id     int64
	text   string
	userId int64
}

func (message *Message) toDTO(userId int64) MessageDTO {
	showId := int64(0)
	if message.id == userId {
		showId = message.id
	}
	return MessageDTO{id: message.id, text: message.text, userId: showId}
}

var seq = int64(0)
var users = make(map[int64]User)
var messages = make(map[int64]Message)

func createMessage(userId int64, text string) (MessageDTO, error) {
	user, userExists := users[userId]
	var m MessageDTO
	if userExists {
		seq = seq + 1
		m := Message{id: seq, text: text, user: &user}
		messages[seq] = m
		return m.toDTO(userId), nil
	}
	return m, errors.New("user does not exist")
}
func getMessages(userId int64) []MessageDTO {
	mArray := make([]MessageDTO, 0, len(messages))
	for _, message := range messages {
		mArray = append(mArray, message.toDTO(userId))
	}
	return mArray
}
func main() {
	users[int64(0)] = User{int64(0), "public"}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.GET("/messages/*userid", func(c *gin.Context) {
		userIdstr := c.Param("userid")
		userId, err := strconv.ParseInt(userIdstr, 10, 64)
		if err != nil {
			userId = int64(0)
		}
		c.JSON(200, gin.H{"message": getMessages(userId)})
	})
	r.PUT("/messages/:userid/:message", func(c *gin.Context) {
		userIdstr := c.Param("userid")
		messageText := c.Param("message")
		userId, err := strconv.ParseInt(userIdstr, 10, 64)
		if err != nil {
			userId = int64(0)
		}
		message, messageError := createMessage(userId, messageText)
		if messageError != nil {
			c.JSON(400, gin.H{"error": messageError.Error()})
		} else {
			c.JSON(200, gin.H{"messageId": message.id, "text": message.text, "by": message.userId})

		}

	})
	r.Run()
}
