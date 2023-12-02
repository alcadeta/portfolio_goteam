//go:build itest

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kxplxn/goteam/internal/api"
	subtaskAPI "github.com/kxplxn/goteam/internal/api/subtask"
	"github.com/kxplxn/goteam/pkg/assert"
	"github.com/kxplxn/goteam/pkg/auth"
	boardTable "github.com/kxplxn/goteam/pkg/dbaccess/board"
	columnTable "github.com/kxplxn/goteam/pkg/dbaccess/column"
	subtaskTable "github.com/kxplxn/goteam/pkg/dbaccess/subtask"
	taskTable "github.com/kxplxn/goteam/pkg/dbaccess/task"
	userTable "github.com/kxplxn/goteam/pkg/dbaccess/user"
	pkgLog "github.com/kxplxn/goteam/pkg/log"
)

// TestSubtaskHandler tests the http.Handler for the subtask API route and
// asserts that it behaves correctly during various execution paths.
func TestSubtaskHandler(t *testing.T) {
	sut := api.NewHandler(
		auth.NewBearerTokenReader(),
		auth.NewJWTValidator(jwtKey),
		map[string]api.MethodHandler{
			http.MethodPatch: subtaskAPI.NewPATCHHandler(
				userTable.NewSelector(db),
				subtaskAPI.NewIDValidator(),
				subtaskTable.NewSelector(db),
				taskTable.NewSelector(db),
				columnTable.NewSelector(db),
				boardTable.NewSelector(db),
				subtaskTable.NewUpdater(db),
				pkgLog.New(),
			),
		},
	)

	t.Run("PATCH", func(t *testing.T) {
		for _, c := range []struct {
			name           string
			id             string
			authFunc       func(*http.Request)
			wantStatusCode int
			assertFunc     func(*testing.T, *http.Response, string)
		}{
			{
				name:           "IDEmpty",
				id:             "",
				authFunc:       addCookieAuth(jwtTeam1Admin),
				wantStatusCode: http.StatusBadRequest,
				assertFunc:     assert.OnResErr("Subtask ID cannot be empty."),
			},
			{
				name:           "IDNotInt",
				id:             "A",
				authFunc:       addCookieAuth(jwtTeam1Admin),
				wantStatusCode: http.StatusBadRequest,
				assertFunc:     assert.OnResErr("Subtask ID must be an integer."),
			},
			{
				name:           "SubtaskNotFound",
				id:             "1001",
				authFunc:       addCookieAuth(jwtTeam1Admin),
				wantStatusCode: http.StatusNotFound,
				assertFunc:     assert.OnResErr("Subtask not found."),
			},
			{
				name:           "BoardWrongTeam",
				id:             "5",
				authFunc:       addCookieAuth(jwtTeam2Admin),
				wantStatusCode: http.StatusForbidden,
				assertFunc: assert.OnResErr(
					"You do not have access to this board.",
				),
			},
			{
				name:           "NotAdmin",
				id:             "5",
				authFunc:       addCookieAuth(jwtTeam1Member),
				wantStatusCode: http.StatusForbidden,
				assertFunc: assert.OnResErr(
					"Only team admins can edit subtasks.",
				),
			},
			{
				name:           "Success",
				id:             "5",
				authFunc:       addCookieAuth(jwtTeam1Admin),
				wantStatusCode: http.StatusOK,
				assertFunc: func(t *testing.T, _ *http.Response, _ string) {
					var isDone bool
					if err := db.QueryRow(
						"SELECT isDone FROM app.subtask WHERE id = 5",
					).Scan(&isDone); err != nil {
						t.Fatal(err)
					}
					assert.True(t.Error, isDone)
				},
			},
		} {
			t.Run(c.name, func(t *testing.T) {
				reqBody, err := json.Marshal(map[string]any{"done": true})
				if err != nil {
					t.Fatal(err)
				}
				r, err := http.NewRequest(
					http.MethodPatch, "?id="+c.id, bytes.NewReader(reqBody),
				)
				if err != nil {
					t.Fatal(err)
				}
				c.authFunc(r)
				w := httptest.NewRecorder()

				sut.ServeHTTP(w, r)
				res := w.Result()

				assert.Equal(t.Error, res.StatusCode, c.wantStatusCode)

				c.assertFunc(t, res, "")
			})
		}

	})
}