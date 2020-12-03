package main

import (
	"errors"
	"fmt"
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
	Id     int64  `json:"id" form:"id" binding:"required"`
	Text   string `json:"text" form:"text" binding:"required"`
	UserId int64  `json:"userId" form:"userId" binding:"required"`
}

func (message *Message) toDTO(userId int64) MessageDTO {
	var showId int64
	if message.user.id == userId {
		showId = userId
	}
	return MessageDTO{Id: message.id, Text: message.text, UserId: showId}
}

const PUBLICUSER = int64(0)

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
func getMessage(userId, messageId int64) (Message, error) {
	message, exists := messages[messageId]
	var m Message
	if exists {
		if user := *message.user; user.id == userId {
			return message, nil
		} else {
			fmt.Println("NOT THE SAME USER", message, userId)
		}
	}
	return m, errors.New(fmt.Sprintf("message %d does not exist", messageId))
}
func deleteMessage(userId, messageId int64) error {
	_, err := getMessage(userId, messageId)
	if err == nil {
		delete(messages, messageId)
		return nil
	}
	return err
}
func editMessage(dto MessageDTO) (MessageDTO, error) {
	message, err := getMessage(dto.UserId, dto.Id)
	var m MessageDTO
	if err != nil {
		return m, err
	}
	message.text = dto.Text
	messages[dto.Id] = message
	return dto, nil
}
func main() {

	users[PUBLICUSER] = User{PUBLICUSER, "public"}
	createMessage(PUBLICUSER, "First !!!!!")
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
		messages := getMessages(userId)
		c.JSON(200, messages)
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
			c.JSON(200, message)

		}

	})
	r.POST("/messages/:messageid/:userid/:message", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("messageid"), 10, 64)
		text := c.Param("message")
		userId, err := strconv.ParseInt(c.Param("userid"), 10, 64)
		if err != nil {
			userId = int64(0)
		}
		message, messageError := editMessage(MessageDTO{id, text, userId})
		if messageError != nil {
			c.JSON(400, gin.H{"error": messageError.Error()})
		} else {
			c.JSON(200, message)

		}

	})
	r.DELETE("/messages/:messageid/:userid/", func(c *gin.Context) {
		messageId, err := strconv.ParseInt(c.Param("messageid"), 10, 64)
		userId, err := strconv.ParseInt(c.Param("userid"), 10, 64)
		if err != nil {
			userId = int64(0)
		}
		messageError := deleteMessage(userId, messageId)
		if messageError == nil {
			c.JSON(200, gin.H{"message": "DELTETED"})

		} else {
			c.JSON(400, gin.H{"error": messageError.Error()})
		}

	})
	r.Run()
}
