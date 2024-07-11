package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	server "github.com/Onnywrite/tinkoff-prod/internal/http-server"
	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage/pg"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tokens.AccessSecret = []byte("$my_%SUPER(n0t-so=MUch)_secret123")
	tokens.RefreshSecret = []byte("$my_%SUPER(n0t-so=MUch)_secret123")
	tokens.AccessTTL = time.Minute
	tokens.RefreshTTL = 240 * time.Hour
	ero.CurrentService = "tinkoff-prod"

	var serverAddress, pgURL string
	flag.StringVar(&serverAddress, "server-address", "", "server port")
	flag.StringVar(&pgURL, "pg-conn", "", "postgres connection string")

	flag.Parse()

	if serverAddress == "" {
		serverAddress = os.Getenv("SERVER_ADDRESS")
		if serverAddress == "" {
			logger.Error("missed SERVER_ADDRESS env (export smth like ':8080')")
			os.Exit(1)
		}
	}

	if pgURL == "" {
		pgURL = os.Getenv("POSTGRES_CONN")
		if pgURL == "" {
			logger.Error("missed POSTGRES_CONN env")
			os.Exit(1)
		}
	}

	db, err := pg.New(pgURL)
	if err != nil {
		logger.Error("server has been stopped", "error", err)
		os.Exit(1)
	}

	s := server.NewServer(serverAddress, db, logger)
	if err = s.Start(); err != nil {
		logger.Error("server has been stopped", "error", err)
	}

	// gracefull shutdown
	shut := make(chan os.Signal, 1)
	signal.Notify(shut, os.Interrupt)
	<-shut
	if err = db.Disconnect(); err != nil {
		logger.Error("could not disconnect from database", "error", err)
	}
	logger.Info("finished")
}
