package board

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/assert"
	"server/dbaccess"
	pkgLog "server/log"
)

// TestPATCHHandler tests the Handle method of PATCHHandler to assert that it
// behaves correctly in all possible scenarios.
func TestPATCHHandler(t *testing.T) {
	log := &pkgLog.FakeErrorer{}
	idValidator := &fakeStringValidator{}
	nameValidator := &fakeStringValidator{}
	boardSelector := &dbaccess.FakeBoardSelector{}
	userBoardSelector := &dbaccess.FakeUserBoardSelector{}
	sut := NewPATCHHandler(
		idValidator, nameValidator, boardSelector, userBoardSelector, log,
	)

	assertOnResErr := func(errMsg string) func(*testing.T, *http.Response) {
		return func(t *testing.T, res *http.Response) {
			var resBody ResBody
			if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
				t.Fatal(err)
			}
			if err := assert.Equal(errMsg, resBody.Error); err != nil {
				t.Error(err)
			}
		}
	}

	assertOnLoggedErr := func(errMsg string) func(*testing.T, *http.Response) {
		return func(t *testing.T, res *http.Response) {
			if err := assert.Equal(errMsg, log.InMessage); err != nil {
				t.Error(err)
			}
		}
	}

	for _, c := range []struct {
		name                        string
		idValidatorOutErr           error
		nameValidatorOutErr         error
		boardSelectorOutErr         error
		userBoardSelectorOutIsAdmin bool
		userBoardSelectorOutErr     error
		wantStatusCode              int
		assertFunc                  func(*testing.T, *http.Response)
	}{
		{
			name:                        "IDValidatorErr",
			idValidatorOutErr:           errors.New("Board ID cannot be empty."),
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         nil,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     nil,
			wantStatusCode:              http.StatusBadRequest,
			assertFunc: assertOnResErr(
				"Board ID cannot be empty.",
			),
		},
		{
			name:                        "NameValidatorErr",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         errors.New("Board name cannot be empty."),
			boardSelectorOutErr:         nil,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     nil,
			wantStatusCode:              http.StatusBadRequest,
			assertFunc: assertOnResErr(
				"Board name cannot be empty.",
			),
		},
		{
			name:                        "BoardNotFound",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         sql.ErrNoRows,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     nil,
			wantStatusCode:              http.StatusNotFound,
			assertFunc:                  assertOnResErr("Board not found."),
		},
		{
			name:                        "BoardSelectorErr",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         sql.ErrConnDone,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     nil,
			wantStatusCode:              http.StatusInternalServerError,
			assertFunc:                  assertOnLoggedErr(sql.ErrConnDone.Error()),
		},
		{
			name:                        "UserDoesNotHaveAccess",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         nil,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     sql.ErrNoRows,
			wantStatusCode:              http.StatusForbidden,
			assertFunc: assertOnResErr(
				"You do not have access to this board.",
			),
		},
		{
			name:                        "UserBoardSelectorErr",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         nil,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     sql.ErrConnDone,
			wantStatusCode:              http.StatusInternalServerError,
			assertFunc:                  assertOnLoggedErr(sql.ErrConnDone.Error()),
		},
		{
			name:                        "UserIsNotAdmin",
			idValidatorOutErr:           nil,
			nameValidatorOutErr:         nil,
			boardSelectorOutErr:         nil,
			userBoardSelectorOutIsAdmin: false,
			userBoardSelectorOutErr:     nil,
			wantStatusCode:              http.StatusForbidden,
			assertFunc: assertOnResErr(
				"Only board admins can edit the board.",
			),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			idValidator.OutErr = c.idValidatorOutErr
			nameValidator.OutErr = c.nameValidatorOutErr
			boardSelector.OutErr = c.boardSelectorOutErr
			userBoardSelector.OutIsAdmin = c.userBoardSelectorOutIsAdmin
			userBoardSelector.OutErr = c.userBoardSelectorOutErr

			reqBody, err := json.Marshal(ReqBody{})
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(
				http.MethodPatch, "", bytes.NewReader(reqBody),
			)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			sut.Handle(w, req, "")
			res := w.Result()

			if err = assert.Equal(
				c.wantStatusCode, res.StatusCode,
			); err != nil {
				t.Error(err)
			}

			c.assertFunc(t, res)
		})
	}
}