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
func TestCRUDMessages(t *testing.T) {

	message := testCreateMessage(t, int64(42), "hello world")
	message2 := testCreateMessage(t, int64(1337), "hello worlds")
	if message.Id == message2.Id {
		t.Errorf("Expected messageId to not match %d and %d", message.Id, message2.Id)
	}
	if foundMessage, messageNotFoundErr := getMessageById(message.Id); messageNotFoundErr != nil {
		t.Errorf("Expected message to be found %d, has error %s", message.Id, messageNotFoundErr.Error())
	} else if foundMessage.Id != message.Id {
		t.Errorf("Expected messageId to match expected %d, actual %d", message.Id, foundMessage.Id)
	}
	testUpdateMessage(t, message)
	testDeleteMessage(t, message)

}

func testUpdateMessage(t *testing.T, message Message) {
	updatedText := "EDIT"
	if updated, updatedError := updateMessage(MessageDTO{Id: message.Id, Text: updatedText, UserId: message.userId}); updatedError != nil {
		t.Errorf("Expected update, but got error %s, %#v", updatedError.Error(), message)
	} else if updated.Id != message.Id {
		t.Errorf("Expected message update, got different ids, expected %d got %d", message.Id, updated.Id)
	} else if updated.Text != updatedText {
		t.Errorf("Expected text update, got same text, expected %s got %s", updatedText, updated.Text)
	}
}
func testDeleteMessage(t *testing.T, message Message) {
	createMessage(message.userId, message.Text)
	nMessages := len(getMessages(0))
	deleteMessageWithId(message.Id)
	lessMessages := len(getMessages(0))
	if nMessages-lessMessages != 1 {
		fmt.Println(getMessages(0))
		t.Errorf("Expected one message to be removed, from %d to %d", nMessages, lessMessages)
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
func TestGetUserByFunctions(t *testing.T) {
	username := "TestGetUserBy"
	user, _ := createUser(username, "password")

	if foundUser, getUserErr := getUserById(user.Id); getUserErr != nil || foundUser.Id != user.Id {
		t.Errorf("Expected to retrieve the same user as created expected %d actual %d", user.Id, foundUser.Id)
	}
	if noUser, noUserErr := getUserById(8080); noUserErr == nil {
		t.Errorf("User does not exist, found %d ", noUser.Id)
	}
	if usernameUser, usernameErr := getUserByUsername(username); usernameErr != nil {
		t.Errorf("User was not found by username %s ", usernameErr.Error())

	} else if usernameUser.Id != user.Id {
		t.Errorf("User was not the expected %d, actual %d ", user.Id, usernameUser.Id)
	}
}

func TestDeleteUser(t *testing.T) {
	user, _ := createUser("testDelete", "password")
	deleteUser(user.Id)
	if removedUser, notFoundError := getUserById(user.Id); notFoundError == nil {
		t.Errorf("User still exists  %#v", removedUser)
	}
}

func testCreateMessage(t *testing.T, userId int64, text string) Message {
	message := createMessage(userId, text)
	if message.userId != userId {
		t.Errorf("Expected userId %d, actual %d", userId, message.userId)
	}
	if message.Text != text {
		t.Errorf("Expected text to be %s, actual %s", text, message.Text)
	}
	return message

}
