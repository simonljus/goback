package main

import (
	"fmt"
	"testing"
)

func TestGetMessages(t *testing.T) {

	if total := getMessages(0); len(total) != 0 {
		t.Errorf("Messages shold be empty")
	}
}
func TestCreateMessage(t *testing.T) {
	dto := testCreateMessage(t, int64(42), "hello world")
	dto2 := testCreateMessage(t, int64(1337), "hello worlds")
	if dto.Id == dto2.Id {
		t.Errorf("Expected messageId to not match %d and %d", dto.Id, dto2.Id)
	}

}
func TestCreateUser(t *testing.T) {
	username := "hi"
	password := "qwerty"
	user, err := createUser(username, password)
	if err != nil {
		t.Errorf("Expected user to be created, got error: %s", err.Error())
	}
	if user.Username != username {
		t.Errorf("Expected username %s, actual %s", username, user.Username)
	}
	if user.passsword != "qwerty" {
		t.Errorf("Expected password to be %s, actual %s", password, user.passsword)
	}
	existingUser, existingError := createUser(username, password)
	if existingError == nil {
		t.Errorf("Expected user to already exist %s,", existingUser.Username)
	}
	otherUser, noError := createUser("qwerty", "hi")
	if noError != nil {
		t.Errorf("Expected user to be created, got error: %s", noError.Error())
	}
	if otherUser.Id == user.Id {
		t.Errorf("Expected messageId to not match %d and %d", user.Id, otherUser.Id)
	}
}
func TestDeleteMessage(t *testing.T) {
	dto := createMessage(42, "hello")
	createMessage(42, "hello")
	nMessages := len(getMessages(0))
	deleteMessageWithId(dto.Id)
	lessMessages := len(getMessages(0))
	if nMessages-lessMessages != 1 {
		fmt.Println(getMessages(0))
		t.Errorf("Expected one message to be removed, from %d to %d", nMessages, lessMessages)
	}
}
func testCreateMessage(t *testing.T, userId int64, text string) MessageDTO {
	dto := createMessage(userId, text)
	if dto.UserId != userId {
		t.Errorf("Expected userId %d, actual %d", userId, dto.UserId)
	}
	if dto.Text != text {
		t.Errorf("Expected text to be %s, actual %s", text, dto.Text)
	}
	return dto

}
