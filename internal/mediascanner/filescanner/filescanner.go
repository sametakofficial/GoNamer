package filescanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
)

var (
	allowedExt = []string{".mkv", ".mp4", ".avi", ".mov", ".flv", ".wmv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp", ".3g2", ".ogv", ".ogg", ".drc", ".gif", ".gifv", ".mng", ".avi", ".mov", ".qt", ".wmv", ".yuv", ".rm", ".rmvb", ".asf", ".amv", ".mp4", ".m4p", ".m4v", ".mpg", ".mp2", ".mpeg", ".mpe", ".mpv", ".mpg", ".mpeg", ".m2v", ".m4v", ".svi", ".3gp", ".3g2", ".mxf", ".roq", ".nsv", ".flv", ".f4v", ".f4p", ".f4a", ".f4b"}
)

type FileScanner struct {
}

func New() mediascanner.MediaScanner {
	return &FileScanner{}
}

func (f *FileScanner) ScanMovies(ctx context.Context, path string, options ...mediascanner.ScanMoviesOptions) (movies []mediascanner.Movie, err error) {
	log := logger.FromContext(ctx)
	var opts mediascanner.ScanMoviesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(ctx, path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			movies = append(movies, parseMovieFileName(ctx, file))
		}
	}
	return
}

func (f *FileScanner) ScanEpisodes(ctx context.Context, path string, options ...mediascanner.ScanEpisodesOptions) (episodes []mediascanner.Episode, err error) {
	log := logger.FromContext(ctx)
	var opts mediascanner.ScanEpisodesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(ctx, path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			ctx = logger.InjectLogger(ctx, log.With("file", file))
			parsed := parseEpisodeFileName(ctx, file, opts.ExcludeUnparsed)
			if parsed.Name == "" {
				continue
			}
			episodes = append(episodes, parsed)
		}
	}
	return
}

func scanDirectory(ctx context.Context, path string, recursive bool) (files []string, err error) {
	log := logger.FromContext(ctx)
	err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.With("error", err).Error("Error accessing path")
			return err
		}

		if !d.IsDir() {
			// Append absolute path
			files = append(files, filePath)
		} else if !recursive && path != filePath {
			return filepath.SkipDir
		}

		return nil
	})
	return
}

func isFileAllowedExt(filename string) bool {
	return slices.Contains(allowedExt, filepath.Ext(filepath.Base(filename)))
}
