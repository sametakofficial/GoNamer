package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nouuu/gonamer/cmd/cli"
	"github.com/nouuu/gonamer/conf"
	"github.com/nouuu/gonamer/internal/cache"
	"github.com/nouuu/gonamer/internal/mediadata/tmdb"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner/filescanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"github.com/pterm/pterm"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()

	initLogger()

	startCli(ctx)
}

func initLogger() {
	logger.SetLoggerLevel(zapcore.InfoLevel)
	logfile, err := os.OpenFile("mediatracker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v\n", err)
		os.Exit(1)
	}

	logger.SetLoggerOutput(zapcore.WriteSyncer(logfile))
}

func startCli(ctx context.Context) {
	log := logger.FromContext(ctx)

	pterm.DefaultHeader.Println("Media Renamer")
	pterm.Print("\n\n")

	pterm.Info.Printfln("Loading configuration...\n")

	config := conf.LoadConfig()

	if config.DryRun {
		pterm.Info.Println("Dry run enabled")
	} else {
		pterm.Warning.Println("Dry run disabled")
	}

	cacheClient, err := cache.NewGoCache(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprintf("Error creating cache client: %v", err))
		log.Fatalf("Error creating cache client: %v", err)
	}

	scanner := filescanner.New()
	movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, cacheClient, tmdb.WithLang("fr-FR"))
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error creating movie client: %v", err))
		log.Fatalf("Error creating movie client: %v", err)
	}
	tvShowClient, err := tmdb.NewTvShowClient(config.TMDBAPIKey, cacheClient, tmdb.WithLang("fr-FR"))
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error creating tv show client: %v", err))
		log.Fatalf("Error creating tv show client: %v", err)
	}

	mediaRenamer := mediarenamer.NewMediaRenamer(movieClient, tvShowClient)

	newCli := cli.NewCli(scanner, mediaRenamer, movieClient, tvShowClient)

	newCli.Run(ctx)
}
