package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/steinfletcher/apitest"
)

func TestGetMessages(t *testing.T) {

	if total := getMessages(0); len(total) != 0 {
		t.Errorf("Messages shold be empty")
	}
}

func TestCRUDMessages(t *testing.T) {

	message := testCreateMessage(t, int64(42), "answer to everything hello world")
	message2 := testCreateMessage(t, int64(1337), "leet hello worlds")
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

func extractCookie(result apitest.Result) *http.Cookie {
	cookie := result.Response.Cookies()[0]
	return cookie
}

func testMeReq(t *testing.T, engine *gin.Engine, cookie *http.Cookie) {
	apitest.New().
		Handler(engine).
		Get("/me").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusOK).
		End()
}
func testCreateMessageReq(t *testing.T, engine *gin.Engine, cookie *http.Cookie, dto MessageDTO) int {
	result := apitest.New().
		Handler(engine).
		Post("/message").
		FormData("message", dto.Text).
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusCreated).
		End()
	id, _ := getIdFromResponse(*result.Response)
	return id
}

func testDeleteMessageReqNoCoookie(t *testing.T, engine *gin.Engine, messageId int) {
	apitest.New().
		Handler(engine).
		Delete(fmt.Sprint("/message/", messageId)).
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
}
func getIdFromResponse(resp http.Response) (int, error) {
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if arr := strings.Split(string(body), `"id":`); len(arr) > 1 {
		idarr := strings.Split(arr[1], ",")
		if id, err := strconv.Atoi(idarr[0]); err != nil {
			return id, err
		} else {
			return id, nil
		}
	}
	return -1, fmt.Errorf("Could not find userid in response body %s", string(body))
}

func createAndSignIn(t *testing.T, engine *gin.Engine, username string, password string) (*http.Cookie, int) {
	signupResult := apitest.New().
		Handler(engine).
		Post("/signup").
		FormData("username", username).
		FormData("password", password).
		Expect(t).
		End()
	result := apitest.New().
		Handler(engine).
		Post("/signin").
		FormData("username", username).
		FormData("password", password).
		Expect(t).
		End()
	id, _ := getIdFromResponse(*signupResult.Response)
	cookie := extractCookie(result)
	return cookie, id
}
func apiCreateMessage(t *testing.T, engine *gin.Engine, cookie *http.Cookie, message string) int {
	result := apitest.New().
		Handler(engine).
		Post("/message").
		FormData("message", message).
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		End()
	id, _ := getIdFromResponse(*result.Response)
	return id
}
func testSignoutReq(t *testing.T, engine *gin.Engine, cookie *http.Cookie) {
	apitest.New().
		Handler(engine).
		Post("/signout").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusNoContent).
		End()
}

func TestGetNotFound(t *testing.T) {
	apitest.New().
		Handler(createEngine()).
		Get("/pathdoesnotexist").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestReqGetMessages(t *testing.T) {
	apitest.New().
		Handler(createEngine()).
		Get("/messages").
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestSignIn(t *testing.T) {
	engine := createEngine()
	username := "testsignin"
	apitest.New().
		Handler(engine).
		Post("/signup").
		FormData("username", username).
		FormData("password", "password").
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(engine).
		Post("/signin").
		Expect(t).
		Status(http.StatusUnauthorized).
		Body(`{"error": "The username and password did not match"}`).
		End()

	apitest.New().
		Handler(engine).
		Post("/signin").
		FormData("username", username).
		FormData("password", "wrongpassword").
		Expect(t).
		Status(http.StatusUnauthorized).
		Body(`{"error": "The username and password did not match"}`).
		End()
	apitest.New().
		Handler(engine).
		Post("/signin").
		FormData("username", username).
		FormData("password", "password").
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestWithoutCookie(t *testing.T) {
	engine := createEngine()
	createAndSignIn(t, engine, "auser", "password")
	apitest.New().
		Handler(engine).
		Get("/me").
		Expect(t).
		Status(http.StatusUnauthorized).
		End()
}

func TestWithCookie(t *testing.T) {
	engine := createEngine()
	cookie, _ := createAndSignIn(t, engine, "auser", "password")
	apitest.New().
		Handler(engine).
		Get("/me").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusOK).
		End()
	apitest.New().
		Handler(engine).
		Post("/signout").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusNoContent).
		End()

}

func TestCreateMessageReq(t *testing.T) {
	engine := createEngine()
	cookie, userId := createAndSignIn(t, engine, "acreateuser", "anotherpassword")
	dto := MessageDTO{Text: "API Create message", UserId: int64(userId)}
	messageId := testCreateMessageReq(t, engine, cookie, dto)
	dto.Id = int64(messageId)
	/*
		apitest.New().
			Handler(engine).
			Get("/messages").
			Cookie(cookie.Name, cookie.Value).
			Expect(t).
			Status(http.StatusOK).
			Body(fmt.Sprintf(`[{"id":%d,"text":"%s","userId":%d}]`, dto.Id, dto.Text, dto.UserId)).
			End()
	*/

	apitest.New().
		Handler(engine).
		Post("/signout").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusNoContent).
		End()

}
func TestDeleteMessageReq(t *testing.T) {
	engine := createEngine()
	cookie, _ := createAndSignIn(t, engine, "adeleteuser", "apassword")
	messageId := apiCreateMessage(t, engine, cookie, "API Delete me")
	testDeleteMessageReqNoCoookie(t, engine, messageId)
	apitest.New().
		Handler(engine).
		Delete("/message/"+fmt.Sprintf("%d", messageId)).
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusNoContent).
		End()

	apitest.New().
		Handler(engine).
		Put("/message/"+fmt.Sprintf("%d", messageId)).
		FormData("message", "THIS WILL NOT UPDATE").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusForbidden).
		End()
	testSignoutReq(t, engine, cookie)
}

func TestUpdateMessageReq(t *testing.T) {
	engine := createEngine()
	cookie, _ := createAndSignIn(t, engine, "aupdateuser", "apassword")
	messageId := apiCreateMessage(t, engine, cookie, "API Update me")
	updatedText := "API UPDATED TEXT"
	apitest.New().
		Handler(engine).
		Put("/message/42").
		FormData("message", updatedText).
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusForbidden).
		End()
	apitest.New().
		Handler(engine).
		Put("/message/"+fmt.Sprintf("%d", messageId)).
		FormData("message", updatedText).
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(engine).
		Get("/messages").
		Cookie(cookie.Name, cookie.Value).
		Expect(t).
		Status(http.StatusOK).
		End()
	testSignoutReq(t, engine, cookie)
}
