package task

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"server/api"
	"server/dbaccess"
	columnTable "server/dbaccess/column"
	taskTable "server/dbaccess/task"
	pkgLog "server/log"
)

// ReqBody defines the request body for requests handled by method handlers.
type ReqBody struct {
	ColumnID      int      `json:"column"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	SubtaskTitles []string `json:"subtasks"`
}

// ResBody defines the response body for requests handled by method handlers.
type ResBody struct {
	Error string `json:"error"`
}

// POSTHandler is an api.MethodHandler that can be used to handle POST task
// requests.
type POSTHandler struct {
	taskTitleValidator    api.StringValidator
	subtaskTitleValidator api.StringValidator
	columnSelector        dbaccess.Selector[columnTable.Record]
	userBoardSelector     dbaccess.RelSelector[bool]
	taskInserter          dbaccess.Inserter[taskTable.Task]
	log                   pkgLog.Errorer
}

// NewPOSTHandler creates and returns a new POSTHandler.
func NewPOSTHandler(
	taskTitleValidator api.StringValidator,
	subtaskTitleValidator api.StringValidator,
	columnSelector dbaccess.Selector[columnTable.Record],
	userBoardSelector dbaccess.RelSelector[bool],
	taskInserter dbaccess.Inserter[taskTable.Task],
	log pkgLog.Errorer,
) *POSTHandler {
	return &POSTHandler{
		taskTitleValidator:    taskTitleValidator,
		subtaskTitleValidator: subtaskTitleValidator,
		columnSelector:        columnSelector,
		userBoardSelector:     userBoardSelector,
		taskInserter:          taskInserter,
		log:                   log,
	}
}

// Handle handles the POST requests sent to the task route.
func (h *POSTHandler) Handle(
	w http.ResponseWriter, r *http.Request, username string,
) {
	var reqBody ReqBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err.Error())
		return
	}

	// Validate task title.
	if err := h.taskTitleValidator.Validate(reqBody.Title); err != nil {
		var errMsg string
		if errors.Is(err, errTitleEmpty) {
			errMsg = "Task title cannot be empty."
		} else if errors.Is(err, errTitleTooLong) {
			errMsg = "Task title cannot be longer than 50 characters."
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(ResBody{
			Error: errMsg,
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err.Error())
		}
		return
	}

	// Validate subtask titles
	for _, title := range reqBody.SubtaskTitles {
		if err := h.subtaskTitleValidator.Validate(title); err != nil {
			var errMsg string
			if errors.Is(err, errTitleEmpty) {
				errMsg = "Subtask title cannot be empty."
			} else if errors.Is(err, errTitleTooLong) {
				errMsg = "Subtask title cannot be longer than 50 characters."
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error(err.Error())
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			if encodeErr := json.NewEncoder(w).Encode(ResBody{
				Error: errMsg,
			}); encodeErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				h.log.Error(err.Error())
			}
			return
		}
	}

	// Get the column from the database with the task's column ID.
	column, err := h.columnSelector.Select(
		strconv.Itoa(reqBody.ColumnID),
	)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		if encodeErr := json.NewEncoder(w).Encode(ResBody{
			Error: "Column not found.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err.Error())
		}
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err.Error())
		return
	}

	// Check if the user is admin on the board the column is associated with.
	isAdmin, err := h.userBoardSelector.Select(
		username, strconv.Itoa(column.BoardID),
	)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)
		if encodeErr := json.NewEncoder(w).Encode(ResBody{
			Error: "You do not have access to this board.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err.Error())
		}
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err.Error())
		return
	}
	if !isAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		if encodeErr := json.NewEncoder(w).Encode(ResBody{
			Error: "Only board admins can create tasks.",
		}); encodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error(err.Error())
		}
		return
	}

	// Insert task and subtasks into the database.
	if err = h.taskInserter.Insert(taskTable.NewTask(
		reqBody.ColumnID,
		reqBody.Title,
		reqBody.Description,
		reqBody.SubtaskTitles,
	)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err.Error())
		return
	}
}