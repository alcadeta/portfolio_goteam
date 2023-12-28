package taskapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/kxplxn/goteam/pkg/cookie"
	"github.com/kxplxn/goteam/pkg/db"
	"github.com/kxplxn/goteam/pkg/db/tasktbl"
	"github.com/kxplxn/goteam/pkg/log"
	"github.com/kxplxn/goteam/pkg/validator"
)

// PostReq defines the body of POST task requests.
type PostReq struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Subtasks     []string `json:"subtasks"`
	BoardID      string   `json:"board"`
	ColumnNumber int      `json:"column"`
	Order        int      `json:"order"`
}

// PostResp defines the body of POST task responses.
type PostResp struct {
	Error string `json:"error"`
}

// PostHandler is an api.MethodHandler that can be used to handle POST requests
// sent to the task route.
type PostHandler struct {
	authDecoder        cookie.Decoder[cookie.Auth]
	titleValidator     validator.String
	subtTitleValidator validator.String
	colNoValidator     validator.Int
	taskInserter       db.Inserter[tasktbl.Task]
	log                log.Errorer
}

// NewPostHandler creates and returns a new POSTHandler.
func NewPostHandler(
	authDecoder cookie.Decoder[cookie.Auth],
	titleValidator validator.String,
	subtTitleValidator validator.String,
	colNoValidator validator.Int,
	taskInserter db.Inserter[tasktbl.Task],
	log log.Errorer,
) *PostHandler {
	return &PostHandler{
		authDecoder:        authDecoder,
		titleValidator:     titleValidator,
		subtTitleValidator: subtTitleValidator,
		colNoValidator:     colNoValidator,
		taskInserter:       taskInserter,
		log:                log,
	}
}

// Handle handles the POST requests sent to the task route.
func (h *PostHandler) Handle(
	w http.ResponseWriter, r *http.Request, _ string,
) {
	// get auth token
	ckAuth, err := r.Cookie(cookie.AuthName)
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusUnauthorized)
		if encodeErr := json.NewEncoder(w).Encode(PostResp{
			Error: "Auth token not found.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
		}
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err)
		return
	}

	// decode auth token
	auth, err := h.authDecoder.Decode(*ckAuth)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encodeErr := json.NewEncoder(w).Encode(PostResp{
			Error: "Invalid auth token.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
		}
		return
	}

	// validate user is admin
	if !auth.IsAdmin {
		w.WriteHeader(http.StatusForbidden)
		if encodeErr := json.NewEncoder(w).Encode(PostResp{
			Error: "Only team admins can create tasks.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
		}
		return
	}

	// decode request
	var req PostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err)
		return
	}

	// validate column ID
	if err := h.colNoValidator.Validate(req.ColumnNumber); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(PostResp{
			Error: "Column number out of bounds.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
		}
		return
	}

	// validate task
	if err := h.titleValidator.Validate(req.Title); err != nil {
		var errMsg string
		if errors.Is(err, validator.ErrEmpty) {
			errMsg = "Task title cannot be empty."
		} else if errors.Is(err, validator.ErrTooLong) {
			errMsg = "Task title cannot be longer than 50 characters."
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		if err = json.NewEncoder(w).Encode(PostResp{
			Error: errMsg,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err)
		}
		return
	}

	// validate subtasks
	var subtasks []tasktbl.Subtask
	for _, title := range req.Subtasks {
		if err := h.subtTitleValidator.Validate(title); err != nil {
			var errMsg string
			if errors.Is(err, validator.ErrEmpty) {
				errMsg = "Subtask title cannot be empty."
			} else if errors.Is(err, validator.ErrTooLong) {
				errMsg = "Subtask title cannot be longer than 50 characters."
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error(err)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			if err = json.NewEncoder(w).Encode(PostResp{
				Error: errMsg,
			}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error(err)
			}
			return
		}
		subtasks = append(subtasks, tasktbl.Subtask{
			Title: title, IsDone: false,
		})
	}

	// insert a new task into the task table - retry up to 3 times for the
	// unlikely event that the generated UUID is a duplicate
	for tries := 0; tries < 3; tries++ {
		id := uuid.NewString()
		if err = h.taskInserter.Insert(r.Context(), tasktbl.NewTask(
			auth.TeamID,
			req.BoardID,
			req.ColumnNumber,
			id,
			req.Title,
			req.Description,
			req.Order,
			subtasks,
		)); errors.Is(err, db.ErrDupKey) {
			continue
		} else if err != nil {
			break
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err)
		return
	}
}
