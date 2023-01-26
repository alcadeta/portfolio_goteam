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
	"server/db"
	"server/log"
)

// TestPOSTHandler tests the Handle method of POSTHandler to assert that it
// behaves correctly in all possible scenarios.
func TestPOSTHandler(t *testing.T) {
	userBoardCounter := &db.FakeCounter{}
	dbBoardInserter := &db.FakeBoardInserter{}
	logger := &log.FakeLogger{}
	sut := NewPostHandler(userBoardCounter, dbBoardInserter, logger)
	sub := "bob123"

	boardInserterErr := errors.New("create board error")

	t.Run(http.MethodPost, func(t *testing.T) {
		for _, c := range []struct {
			name                   string
			reqBody                ReqBody
			userBoardCounterOutRes int
			userBoardCounterOutErr error
			boardInserterOutErr    error
			wantStatusCode         int
			wantErr                string
		}{
			{
				name:                   "BoardNameNil",
				reqBody:                ReqBody{},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: nil,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameEmpty,
			},
			{
				name:                   "BoardNameEmpty",
				reqBody:                ReqBody{Name: ""},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: nil,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameEmpty,
			},
			{
				name: "BoardNameTooLong",
				reqBody: ReqBody{
					Name: "boardyboardsyboardkyboardishboardxyza",
				},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: nil,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errNameTooLong,
			},
			{
				name:                   "UserBoardCounterErr",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: sql.ErrConnDone,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusInternalServerError,
				wantErr:                errMaxBoards,
			},
			{
				name:                   "MaxBoardsCreated",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 3,
				userBoardCounterOutErr: nil,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusBadRequest,
				wantErr:                errMaxBoards,
			},
			{
				name:                   "BoardInserterErr",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: sql.ErrNoRows,
				boardInserterOutErr:    boardInserterErr,
				wantStatusCode:         http.StatusInternalServerError,
				wantErr:                boardInserterErr.Error(),
			},
			{
				name:                   "Success",
				reqBody:                ReqBody{Name: "someboard"},
				userBoardCounterOutRes: 0,
				userBoardCounterOutErr: sql.ErrNoRows,
				boardInserterOutErr:    nil,
				wantStatusCode:         http.StatusOK,
				wantErr:                "",
			},
		} {
			t.Run(c.name, func(t *testing.T) {
				userBoardCounter.OutRes = c.userBoardCounterOutRes
				userBoardCounter.OutErr = c.userBoardCounterOutErr
				dbBoardInserter.OutErr = c.boardInserterOutErr

				reqBodyJSON, err := json.Marshal(c.reqBody)
				if err != nil {
					t.Fatal(err)
				}
				req, err := http.NewRequest(
					http.MethodPost, "/board", bytes.NewReader(reqBodyJSON),
				)
				if err != nil {
					t.Fatal(err)
				}

				w := httptest.NewRecorder()

				sut.Handle(w, req, sub)

				if err = assert.Equal(
					c.wantStatusCode, w.Result().StatusCode,
				); err != nil {
					t.Error(err)
				}

				// if 400 is expected - there must be a validation error in
				// response body
				if c.wantStatusCode == http.StatusBadRequest {
					resBody := ResBody{}
					if err := json.NewDecoder(w.Result().Body).Decode(
						&resBody,
					); err != nil {
						t.Error(err)
					}

					if err := assert.Equal(
						c.wantErr, resBody.Error,
					); err != nil {
						t.Error(err)
					}
				}

				// DEPENDENCY-INPUT-BASED ASSERTIONS

				if c.wantStatusCode == http.StatusInternalServerError {
					errFound := false
					for _, err := range []error{
						c.userBoardCounterOutErr, c.boardInserterOutErr,
					} {
						if err != nil && err != sql.ErrNoRows {
							errFound = true
							if err := assert.Equal(
								log.LevelError, logger.InLevel,
							); err != nil {
								t.Error(err)
							}
							if err := assert.Equal(err.Error(), logger.InMessage); err != nil {
								t.Error(err)
							}
						}
					}
					if !errFound {
						t.Errorf(
							"c.wantStatusCode was %d but no errors were logged.",
							http.StatusInternalServerError,
						)
					}
					return
				}

				// if max boards is not reached, board creator must be called
				if c.userBoardCounterOutRes >= maxBoards ||
					c.reqBody.Name == "" ||
					len(c.reqBody.Name) > maxNameLength {
					return
				}
				if err := assert.Equal(
					db.NewBoard(c.reqBody.Name, sub),
					dbBoardInserter.InBoard,
				); err != nil {
					t.Error(err)
				}
			})
		}
	})
}
