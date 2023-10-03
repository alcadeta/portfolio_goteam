package main

import (
	"database/sql"
	"net/http"
	"os"

	"server/api"
	"server/api/board"
	"server/api/login"
	"server/api/register"
	"server/auth"
	"server/dbaccess"
	pkgLog "server/log"
	"server/midware"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	// Create a log for the app.
	log := pkgLog.New()

	// Load environment variables from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the necessary env vars were set.
	env := newEnv()
	if err := env.validate(); err != nil {
		log.Fatal(err.Error())
		os.Exit(2)
	}

	// Create dependencies that are shared by multiple handlers.
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
	userSelector := dbaccess.NewUserSelector(db)

	// Register handlers for API routes.
	mux := http.NewServeMux()

	mux.Handle("/register", register.NewHandler(
		register.NewValidator(
			register.NewUsernameValidator(),
			register.NewPasswordValidator(),
		),
		userSelector,
		register.NewPasswordHasher(),
		dbaccess.NewUserInserter(db),
		jwtGenerator,
		log,
	))

	mux.Handle("/login", login.NewHandler(
		login.NewValidator(),
		userSelector,
		login.NewPasswordComparator(),
		jwtGenerator,
		log,
	))

	mux.Handle("/board", board.NewHandler(
		auth.NewBearerTokenReader(),
		auth.NewJWTValidator(env.JWTKey),
		map[string]api.MethodHandler{
			http.MethodPost: board.NewPOSTHandler(
				board.NewNameValidator(),
				dbaccess.NewUserBoardCounter(db),
				dbaccess.NewBoardInserter(db),
				log,
			),
			http.MethodDelete: board.NewDELETEHandler(
				board.NewIDValidator(),
				dbaccess.NewUserBoardSelector(db),
				dbaccess.NewBoardDeleter(db),
				log,
			),
		},
	))

	// Set up CORS.
	handler := midware.NewCORS(mux, env.ClientOrigin)

	// Serve the app using the ServeMux.
	log.Info("running server at port " + env.Port)
	if err := http.ListenAndServe(":"+env.Port, handler); err != nil {
		log.Fatal(err.Error())
		os.Exit(5)
	}
}
