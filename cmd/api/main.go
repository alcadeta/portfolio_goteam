package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/kxplxn/goteam/internal/api"
	boardAPI "github.com/kxplxn/goteam/internal/api/board"
	columnAPI "github.com/kxplxn/goteam/internal/api/column"
	loginAPI "github.com/kxplxn/goteam/internal/api/login"
	registerAPI "github.com/kxplxn/goteam/internal/api/register"
	subtaskAPI "github.com/kxplxn/goteam/internal/api/subtask"
	taskAPI "github.com/kxplxn/goteam/internal/api/task"
	"github.com/kxplxn/goteam/pkg/auth"
	boardTable "github.com/kxplxn/goteam/pkg/dbaccess/board"
	columnTable "github.com/kxplxn/goteam/pkg/dbaccess/column"
	subtaskTable "github.com/kxplxn/goteam/pkg/dbaccess/subtask"
	taskTable "github.com/kxplxn/goteam/pkg/dbaccess/task"
	teamTable "github.com/kxplxn/goteam/pkg/dbaccess/team"
	userTable "github.com/kxplxn/goteam/pkg/dbaccess/user"
	pkgLog "github.com/kxplxn/goteam/pkg/log"
)

func main() {
	// Create a logger for the app.
	log := pkgLog.New()

	// Load environment variables from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the necessary env vars were set.
	env := api.NewEnv()
	if err := env.Validate(); err != nil {
		log.Fatal(err.Error())
		os.Exit(2)
	}

	// Create dependencies that are used by multiple handlers.
	db, err := sql.Open("postgres", env.DBConnStr)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(3)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err.Error())
		os.Exit(4)
	}
	jwtGenerator := auth.NewJWTGenerator(env.JWTKey)
	bearerTokenReader := auth.NewBearerTokenReader()
	jwtValidator := auth.NewJWTValidator(env.JWTKey)
	userSelector := userTable.NewSelector(db)
	columnSelector := columnTable.NewSelector(db)

	// Register handlers for API routes.
	mux := http.NewServeMux()

	teamSelectorByInvCode := teamTable.NewSelectorByInvCode(db)
	mux.Handle("/register", api.NewHandler(nil, nil,
		map[string]api.MethodHandler{
			http.MethodPost: registerAPI.NewPOSTHandler(
				registerAPI.NewUserValidator(
					registerAPI.NewUsernameValidator(),
					registerAPI.NewPasswordValidator(),
				),
				registerAPI.NewInviteCodeValidator(),
				teamSelectorByInvCode,
				userSelector,
				registerAPI.NewPasswordHasher(),
				userTable.NewInserter(db),
				jwtGenerator,
				log,
			),
		},
	))

	mux.Handle("/login", api.NewHandler(nil, nil,
		map[string]api.MethodHandler{
			http.MethodPost: loginAPI.NewPOSTHandler(
				loginAPI.NewValidator(),
				userSelector,
				loginAPI.NewPasswordComparator(),
				jwtGenerator,
				log,
			),
		},
	))

	boardIDValidator := boardAPI.NewIDValidator()
	boardNameValidator := boardAPI.NewNameValidator()
	boardSelector := boardTable.NewSelector(db)
	boardInserter := boardTable.NewInserter(db)
	mux.Handle("/board", api.NewHandler(
		bearerTokenReader,
		jwtValidator,
		map[string]api.MethodHandler{
			http.MethodGet: boardAPI.NewGETHandler(
				userSelector,
				boardInserter,
				boardIDValidator,
				boardTable.NewRecursiveSelector(db),
				teamTable.NewSelector(db),
				userTable.NewSelectorByTeamID(db),
				boardTable.NewSelectorByTeamID(db),
				log,
			),
			http.MethodPost: boardAPI.NewPOSTHandler(
				userSelector,
				boardNameValidator,
				boardTable.NewCounter(db),
				boardInserter,
				log,
			),
			http.MethodDelete: boardAPI.NewDELETEHandler(
				userSelector,
				boardIDValidator,
				boardSelector,
				boardTable.NewDeleter(db),
				log,
			),
			http.MethodPatch: boardAPI.NewPATCHHandler(
				userSelector,
				boardIDValidator,
				boardNameValidator,
				boardSelector,
				boardTable.NewUpdater(db),
				log,
			),
		},
	))

	mux.Handle("/column", api.NewHandler(
		bearerTokenReader,
		jwtValidator,
		map[string]api.MethodHandler{
			http.MethodPatch: columnAPI.NewPATCHHandler(
				userSelector,
				columnAPI.NewIDValidator(),
				columnSelector,
				boardSelector,
				columnTable.NewUpdater(db),
				log,
			),
		},
	))

	taskIDValidator := taskAPI.NewIDValidator()
	taskTitleValidator := taskAPI.NewTitleValidator()
	taskSelector := taskTable.NewSelector(db)
	mux.Handle("/task", api.NewHandler(
		bearerTokenReader,
		jwtValidator,
		map[string]api.MethodHandler{
			http.MethodPost: taskAPI.NewPOSTHandler(
				userSelector,
				taskTitleValidator,
				taskTitleValidator,
				columnSelector,
				boardSelector,
				taskTable.NewInserter(db),
				log,
			),
			http.MethodPatch: taskAPI.NewPATCHHandler(
				userSelector,
				taskIDValidator,
				taskTitleValidator,
				taskTitleValidator,
				taskSelector,
				columnSelector,
				boardSelector,
				taskTable.NewUpdater(db),
				log,
			),
			http.MethodDelete: taskAPI.NewDELETEHandler(
				userSelector,
				taskIDValidator,
				taskSelector,
				columnSelector,
				boardSelector,
				taskTable.NewDeleter(db),
				log,
			),
		},
	))

	mux.Handle("/subtask", api.NewHandler(
		bearerTokenReader,
		jwtValidator,
		map[string]api.MethodHandler{
			http.MethodPatch: subtaskAPI.NewPATCHHandler(
				userSelector,
				subtaskAPI.NewIDValidator(),
				subtaskTable.NewSelector(db),
				taskSelector,
				columnSelector,
				boardSelector,
				subtaskTable.NewUpdater(db),
				pkgLog.New(),
			),
		},
	))

	// Serve the app using the ServeMux.
	log.Info("running server at port " + env.Port)
	if err := http.ListenAndServe(":"+env.Port, mux); err != nil {
		log.Fatal(err.Error())
		os.Exit(5)
	}
}