package filescanner

import (
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/nouuu/mediatracker/internal/logger"
	"github.com/nouuu/mediatracker/internal/mediascanner"
)

var (
	log        = logger.GetLogger()
	allowedExt = []string{".mkv", ".mp4", ".avi", ".mov", ".flv", ".wmv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp", ".3g2", ".ogv", ".ogg", ".drc", ".gif", ".gifv", ".mng", ".avi", ".mov", ".qt", ".wmv", ".yuv", ".rm", ".rmvb", ".asf", ".amv", ".mp4", ".m4p", ".m4v", ".mpg", ".mp2", ".mpeg", ".mpe", ".mpv", ".mpg", ".mpeg", ".m2v", ".m4v", ".svi", ".3gp", ".3g2", ".mxf", ".roq", ".nsv", ".flv", ".f4v", ".f4p", ".f4a", ".f4b"}
)

type FileScanner struct {
}

func New() mediascanner.MediaScanner {
	return &FileScanner{}
}

func (f *FileScanner) ScanMovies(path string, options ...mediascanner.ScanMoviesOptions) (movies []mediascanner.Movie, err error) {
	var opts mediascanner.ScanMoviesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			movies = append(movies, parseMovieFileName(file))
		}
	}
	return
}

func (f *FileScanner) ScanEpisodes(path string, options ...mediascanner.ScanEpisodesOptions) (episodes []mediascanner.Episode, err error) {
	var opts mediascanner.ScanEpisodesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			episodes = append(episodes, parseEpisodeFileName(file))
		}
	}
	return
}

func scanDirectory(path string, recursive bool) (files []string, err error) {
	err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.With("error", err).Error("Error accessing path")
			return err
		}

		if !d.IsDir() {
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
