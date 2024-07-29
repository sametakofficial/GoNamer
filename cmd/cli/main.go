package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/mediadata/tmdb"
	"github.com/nouuu/mediatracker/internal/mediarenamer"
	"github.com/nouuu/mediatracker/internal/mediascanner/filescanner"
	"github.com/nouuu/mediatracker/pkg/logger"
	"github.com/pterm/pterm"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger.SetLoggerLevel(zapcore.InfoLevel)
	logfile, err := os.OpenFile("mediatracker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v\n", err)
		os.Exit(1)
	}

	logger.SetLoggerOutput(zapcore.WriteSyncer(logfile))
	ctx := context.Background()
	log := logger.FromContext(ctx)

	pterm.DefaultHeader.Println("Media Renamer")
	pterm.Print("\n\n")

	pterm.Info.Printfln("Loading configuration...\n")

	config := conf.LoadConfig()

	scanner := filescanner.New()
	movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error creating movie client: %v", err))
		log.Fatalf("Error creating movie client: %v", err)
	}

	mediaRenamer := mediarenamer.NewMediaRenamer(movieClient)

	cli := NewCli(scanner, mediaRenamer, movieClient)

	cli.Run(ctx)
}
