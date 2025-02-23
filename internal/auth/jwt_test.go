package auth

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	newuser := uuid.New()
	token, err := MakeJWT(newuser, "myubersecret", time.Second*10)
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	fmt.Println(token)
}

func TestValidateJWT(t *testing.T) {
	user1 := uuid.New()
	user2 := uuid.New()

	token1, _ := MakeJWT(user1, "myubersecret", time.Second*10)
	token2, _ := MakeJWT(user2, "myubersecret", time.Millisecond*10)
	time.Sleep(time.Second)

	user1uuid, err := ValidateJWT(token1, "myubersecret")
	if err != nil {
		t.Logf("Failed to validate token: %s", err)
		t.Fail()
	} else if user1uuid != user1 {
		t.Logf("UUID does not match token. Got: %s, want: %s", user1uuid.String(), user1.String())
		t.Fail()

	}

	_, err = ValidateJWT(token2, "myubersecret")
	if err == nil {
		t.Log("Got no error when expecting error")
		t.Fail()
	} else {
		t.Logf("Got expected error: %s", err)
	}
}

func TestGetBearerToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	newuser := uuid.New()
	token, _ := MakeJWT(newuser, "myubersecret", time.Second*300)
	req.Header.Add("Authorization", "Bearer "+token)
	result, err := GetBearerToken(req.Header)
	if err != nil {
		t.Logf("Got error %s", err)
		t.Fail()
		return
	}
	if result == token {
		t.Log("Token successfully extracted")
	} else {
		t.Logf("Extracted %s, expected %s", result, token)
	}
}
