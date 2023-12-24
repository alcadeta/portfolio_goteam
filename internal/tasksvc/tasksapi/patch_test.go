//go:build utest

package tasksapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kxplxn/goteam/pkg/api"
	"github.com/kxplxn/goteam/pkg/assert"
	"github.com/kxplxn/goteam/pkg/cookie"
	"github.com/kxplxn/goteam/pkg/db"
	"github.com/kxplxn/goteam/pkg/db/tasktbl"
	"github.com/kxplxn/goteam/pkg/log"
)

func TestPatchHandler(t *testing.T) {
	authDecoder := &cookie.FakeDecoder[cookie.Auth]{}
	stateDecoder := &cookie.FakeDecoder[cookie.State]{}
	colNoVdtor := &api.FakeIntValidator{}
	tasksUpdater := &db.FakeUpdater[[]tasktbl.Task]{}
	stateEncoder := &cookie.FakeEncoder[cookie.State]{}
	log := &log.FakeErrorer{}
	sut := NewPatchHandler(
		authDecoder,
		stateDecoder,
		colNoVdtor,
		tasksUpdater,
		stateEncoder,
		log,
	)

	for _, c := range []struct {
		name             string
		authToken        string
		errDecodeAuth    error
		authDecoded      cookie.Auth
		stateToken       string
		errDecodeState   error
		stateDecoded     cookie.State
		errValidateColNo error
		errUpdateTasks   error
		errEncodeState   error
		outState         http.Cookie
		wantStatus       int
		assertFunc       func(*testing.T, *http.Response, []any)
	}{
		{
			name:             "NoAuth",
			authToken:        "",
			errDecodeAuth:    nil,
			authDecoded:      cookie.Auth{},
			stateToken:       "",
			errDecodeState:   nil,
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusUnauthorized,
			assertFunc:       assert.OnRespErr("Auth token not found."),
		},
		{
			name:             "ErrDecodeAuth",
			authToken:        "nonempty",
			errDecodeAuth:    errors.New("decode auth failed"),
			authDecoded:      cookie.Auth{},
			stateToken:       "",
			errDecodeState:   nil,
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusUnauthorized,
			assertFunc:       assert.OnRespErr("Invalid auth token."),
		},
		{
			name:             "NotAdmin",
			authToken:        "nonempty",
			errDecodeAuth:    nil,
			authDecoded:      cookie.Auth{IsAdmin: false},
			stateToken:       "",
			errDecodeState:   nil,
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusForbidden,
			assertFunc: assert.OnRespErr(
				"Only team admins can edit tasks.",
			),
		},
		{
			name:             "NoState",
			authToken:        "nonempty",
			errDecodeAuth:    nil,
			authDecoded:      cookie.Auth{IsAdmin: true},
			stateToken:       "",
			errDecodeState:   nil,
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusBadRequest,
			assertFunc:       assert.OnRespErr("State token not found."),
		},
		{
			name:             "ErrDecodeState",
			authToken:        "nonempty",
			errDecodeAuth:    nil,
			authDecoded:      cookie.Auth{IsAdmin: true},
			stateToken:       "nonempty",
			errDecodeState:   errors.New("decode state failed"),
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusBadRequest,
			assertFunc:       assert.OnRespErr("Invalid state token."),
		},
		{
			name:             "NoAccess",
			authToken:        "nonempty",
			errDecodeAuth:    nil,
			authDecoded:      cookie.Auth{IsAdmin: true},
			stateToken:       "nonempty",
			errDecodeState:   nil,
			stateDecoded:     cookie.State{},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusBadRequest,
			assertFunc:       assert.OnRespErr("Invalid task ID."),
		},
		{
			name:           "ColNoInvalid",
			authToken:      "nonempty",
			errDecodeAuth:  nil,
			authDecoded:    cookie.Auth{IsAdmin: true},
			stateToken:     "nonempty",
			errDecodeState: nil,
			stateDecoded: cookie.State{Boards: []cookie.Board{{
				ID:      "1",
				Columns: []cookie.Column{{Tasks: []cookie.Task{{ID: "taskid"}}}},
			}}},
			errValidateColNo: errors.New("err validate column number"),
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusBadRequest,
			assertFunc:       assert.OnRespErr("Invalid column number."),
		},
		{
			name:           "TaskNotFound",
			authToken:      "nonempty",
			errDecodeAuth:  nil,
			authDecoded:    cookie.Auth{IsAdmin: true, TeamID: "1"},
			stateToken:     "nonempty",
			errDecodeState: nil,
			stateDecoded: cookie.State{Boards: []cookie.Board{{
				ID:      "1",
				Columns: []cookie.Column{{Tasks: []cookie.Task{{ID: "taskid"}}}},
			}}},
			errValidateColNo: nil,
			errUpdateTasks:   db.ErrNoItem,
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusNotFound,
			assertFunc:       assert.OnRespErr("Task not found."),
		},
		{
			name:           "ErrUpdateTasks",
			authToken:      "nonempty",
			errDecodeAuth:  nil,
			authDecoded:    cookie.Auth{IsAdmin: true, TeamID: "1"},
			stateToken:     "nonempty",
			errDecodeState: nil,
			stateDecoded: cookie.State{Boards: []cookie.Board{{
				ID:      "1",
				Columns: []cookie.Column{{Tasks: []cookie.Task{{ID: "taskid"}}}},
			}}},
			errValidateColNo: nil,
			errUpdateTasks:   errors.New("update tasks failed"),
			errEncodeState:   nil,
			outState:         http.Cookie{},
			wantStatus:       http.StatusInternalServerError,
			assertFunc:       assert.OnLoggedErr("update tasks failed"),
		},
		{
			name:           "ErrEncodeState",
			authToken:      "nonempty",
			errDecodeAuth:  nil,
			authDecoded:    cookie.Auth{IsAdmin: true, TeamID: "1"},
			stateToken:     "nonempty",
			errDecodeState: nil,
			stateDecoded: cookie.State{Boards: []cookie.Board{{
				ID:      "1",
				Columns: []cookie.Column{{Tasks: []cookie.Task{{ID: "taskid"}}}},
			}}},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   errors.New("encode state failed"),
			outState:         http.Cookie{},
			wantStatus:       http.StatusInternalServerError,
			assertFunc:       assert.OnLoggedErr("encode state failed"),
		},
		{
			name:           "OK",
			authToken:      "nonempty",
			errDecodeAuth:  nil,
			authDecoded:    cookie.Auth{IsAdmin: true, TeamID: "1"},
			stateToken:     "nonempty",
			errDecodeState: nil,
			stateDecoded: cookie.State{Boards: []cookie.Board{{
				ID:      "1",
				Columns: []cookie.Column{{Tasks: []cookie.Task{{ID: "taskid"}}}},
			}}},
			errValidateColNo: nil,
			errUpdateTasks:   nil,
			errEncodeState:   nil,
			outState:         http.Cookie{Name: "foo", Value: "bar"},
			wantStatus:       http.StatusOK,
			assertFunc: func(t *testing.T, resp *http.Response, _ []any) {
				ck := resp.Cookies()[0]
				assert.Equal(t.Error, ck.Name, "foo")
				assert.Equal(t.Error, ck.Value, "bar")
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			authDecoder.Res = c.authDecoded
			authDecoder.Err = c.errDecodeAuth
			stateDecoder.Res = c.stateDecoded
			stateDecoder.Err = c.errDecodeState
			colNoVdtor.Err = c.errValidateColNo
			tasksUpdater.Err = c.errUpdateTasks
			stateEncoder.Err = c.errEncodeState
			stateEncoder.Res = c.outState
			w := httptest.NewRecorder()
			r := httptest.NewRequest("", "/", strings.NewReader(`[{
                "id": "taskid",
                "order": 3,
                "column": 0
            }]`))
			if c.authToken != "" {
				r.AddCookie(&http.Cookie{
					Name: "auth-token", Value: c.authToken,
				})
			}
			if c.stateToken != "" {
				r.AddCookie(&http.Cookie{
					Name: "state-token", Value: c.stateToken,
				})
			}

			sut.Handle(w, r, "")

			resp := w.Result()
			assert.Equal(t.Error, resp.StatusCode, c.wantStatus)
			c.assertFunc(t, resp, log.Args)
		})
	}
}
