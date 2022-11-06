package register

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kxplxn/goteam/server-v2/assert"
)

func TestRegister(t *testing.T) {
	t.Run("UsernameValidation", func(t *testing.T) {
		const (
			empty       = "Username cannot be empty."
			tooShort    = "Username cannot be shorter than 5 characters."
			tooLong     = "Username cannot be longer than 15 characters."
			invalidChar = "Username can contain only letters (a-z/A-Z) and digits (0-9)."
			digitStart  = "Username can start only with a letter (a-z/A-Z)."
		)
		for _, c := range []struct {
			name     string
			username string
			errs     []string
		}{
			{name: "Empty", username: "", errs: []string{empty}},
			{name: "TooShort", username: "bob1", errs: []string{tooShort}},
			{name: "TooLong", username: "bobobobobobobobob", errs: []string{tooLong}},
			{name: "InvalidCharacter", username: "bobob!", errs: []string{invalidChar}},
			{name: "DigitStart", username: "1bobob", errs: []string{digitStart}},
			{name: "TooShort_InvalidCharacter", username: "bob!", errs: []string{tooShort, invalidChar}},
			{name: "TooShort_DigitStart", username: "1bob", errs: []string{tooShort, digitStart}},
			{name: "TooLong_InvalidCharacter", username: "bobobobobobobobo!", errs: []string{tooLong, invalidChar}},
			{name: "TooLong_DigitStart", username: "1bobobobobobobobo", errs: []string{tooLong, digitStart}},
			{name: "InvalidCharacter_DigitStart", username: "1bob!", errs: []string{invalidChar, digitStart}},
			{name: "TooShort_InvalidCharacter_DigitStart", username: "1bo!", errs: []string{tooShort, invalidChar, digitStart}},
			{name: "TooLong_InvalidCharacter_DigitStart", username: "1bobobobobobobob!", errs: []string{tooLong, invalidChar, digitStart}},
		} {
			t.Run(c.name, func(t *testing.T) {
				// arrange
				req, err := http.NewRequest("POST", "/register", strings.NewReader(fmt.Sprintf(`{
					"username": "%s", 
					"password": "SecureP4ss?", 
					"referrer": ""
				}`, c.username)))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler := NewHandler()

				// act
				handler.ServeHTTP(w, req)

				// assert
				res := w.Result()
				gotStatusCode, wantStatusCode := res.StatusCode, http.StatusBadRequest
				if gotStatusCode != wantStatusCode {
					t.Logf("\nwant: %d\ngot: %d", http.StatusBadRequest, res.StatusCode)
					t.Fail()
				}
				resBody := &ResBody{}
				if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
					t.Fatal(err)
				}
				gotErr := resBody.Errs.Username
				if !assert.EqualArr(gotErr, c.errs) {
					t.Logf("\nwant: %+v\ngot: %+v", c.errs, gotErr)
					t.Fail()
				}
			})
		}
	})

	t.Run("PasswordValidation", func(t *testing.T) {
		for _, c := range []struct {
			caseName string
			password string
			wantErr  []string
		}{
			{
				caseName: "Empty",
				password: "",
				wantErr:  []string{"Password cannot be empty."},
			},
			{
				caseName: "TooShort",
				password: "mypassw",
				wantErr:  []string{"Password cannot be shorter than 5 characters."},
			},
			{
				caseName: "TooLong",
				password: "mypasswordwhichislongandimeanreallylongforsomereasonohiknowwhytbh",
				wantErr:  []string{"Password cannot be longer than 64 characters."},
			},
			{
				caseName: "NoLowercase",
				password: "MYALLUPPERPASSWORD",
				wantErr:  []string{"Password must contain a lowercase letter (a-z)."},
			},
			{
				caseName: "NoUppercase",
				password: "myalllowerpassword",
				wantErr:  []string{"Password must contain an uppercase letter (A-Z)."},
			},
			{
				caseName: "NoDigits",
				password: "myNOdigitPASSWORD",
				wantErr:  []string{"Password must contain a digit (0-9)."},
			},
			{
				caseName: "NoSymbols",
				password: "myNOsymbolP4SSWORD",
				wantErr: []string{
					"Password must contain one of the following special characters: " +
						"! \" # $ % & ' ( ) * + , - . / : ; < = > ? [ \\ ] ^ _ ` { | } ~.",
				},
			},
			{
				caseName: "HasSpaces",
				password: "my SP4CED p4ssword",
				wantErr:  []string{"Password cannot contain spaces."},
			},
			{
				caseName: "NonASCII",
				password: "myNØNÅSCÎÎp4ssword",
				wantErr: []string{
					"Password can contain only letters (a-z/A-Z), digits (0-9), " +
						"and the following special characters: " +
						"! \" # $ % & ' ( ) * + , - . / : ; < = > ? [ \\ ] ^ _ ` { | } ~.",
				},
			},
		} {
			t.Run(c.caseName, func(t *testing.T) {
				// arrange
				req, err := http.NewRequest("POST", "/register", strings.NewReader(fmt.Sprintf(`{
					"username": "mynameisbob", 
					"password": "%s", 
					"referrer": ""
				}`, c.password)))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler := NewHandler()

				// act
				handler.ServeHTTP(w, req)

				// assert
				res := w.Result()
				gotStatusCode, wantStatusCode := res.StatusCode, http.StatusBadRequest
				if gotStatusCode != wantStatusCode {
					t.Logf("\nwant: %d\ngot: %d", wantStatusCode, gotStatusCode)
					t.Fail()
				}
				resBody := &ResBody{}
				if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
					t.Fatal(err)
				}
				gotErr := resBody.Errs.Password
				if !assert.EqualArr(gotErr, c.wantErr) {
					t.Logf("\nwant: %+v\ngot: %+v", c.wantErr, gotErr)
					t.Fail()
				}
			})
		}
	})
}
