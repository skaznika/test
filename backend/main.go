package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var UPLOAD_DIR = "/app/uploads"
var LT_ENDPOINT = "translate:5000"
var ASR_ENDPOINT = "whisper-api:8000"

func initEnv() {
	// Load .env file
	Lggr.Info().Msg("Loading .env file.")
	err := godotenv.Load()
	if err != nil {
		Lggr.Info().Msg("No .env file found. Using environment variables or default.")
	}

	if os.Getenv("UPLOAD_DIR") != "" {
		UPLOAD_DIR = os.Getenv("UPLOAD_DIR")
	}
	if os.Getenv("ASR_ENDPOINT") != "" {
		ASR_ENDPOINT = os.Getenv("ASR_ENDPOINT")
	}
	if os.Getenv("LT_ENDPOINT") != "" {
		LT_ENDPOINT = os.Getenv("LT_ENDPOINT")
	}

	Lggr.Info().Msg("Current ENV:")
	Lggr.Info().Msgf("- UPLOAD_DIR: %v", UPLOAD_DIR)
	Lggr.Info().Msgf("- ASR_ENDPOINT: %v", ASR_ENDPOINT)
	Lggr.Info().Msgf("- LT_ENDPOINT: %v\n", LT_ENDPOINT)
}

var Lggr zerolog.Logger

func main() {
	// Flags
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	// Logging
	Lggr = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Load environment variables
	initEnv()

	// Database
	Lggr.Info().Msg("Initialising Database.")
	InitDB()
	Migrate()

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Tasks
	r.Post("/asr", ProcessFrom)
	r.Get("/translate/{id}/{target_language}", TranslateTask)

	// Jobs
	r.Get("/jobs", GetJobs)
	r.Get("/jobs/{id}", GetJobById)
	r.Get("/jobs/delete/{id}", DeleteJob)
	r.Post("/jobs/edit", EditJob)

	// Subtitles
	r.Get("/download/{id}/{format}", GetSubs)
	r.Get("/download/translation/{id}/{format}", GetSubs)

	// Files
	r.Get("/file/{id}", GetFile)

	// Others
	r.Get("/languages", GetLanguages)

	Lggr.Info().Msg("Starting server on port 3000...")
	go MonitorJobs()
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		Lggr.Fatal().Err(err).Msg("Failed to start server")
	}
}
