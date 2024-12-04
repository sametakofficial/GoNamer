package mediarenamer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediascanner"
)

type Field string

const (
	FieldName         Field = "{name}"
	FieldYear         Field = "{year}"
	FieldDate         Field = "{date}"
	FieldExt          Field = "{extension}"
	FieldSeason       Field = "{season}"
	FieldEpisode      Field = "{episode}"
	FieldEpisodeTitle Field = "{episode_title}"
)

func GenerateMovieFilename(pattern string, movie mediadata.Movie, fileMovie mediascanner.Movie) string {
	//return fmt.Sprintf("%s - %s%s", movie.Title, movie.Year, fileMovie.Extension)
	filename := pattern
	filename = replaceField(filename, FieldName, movie.Title)
	filename = replaceField(filename, FieldYear, movie.Year)
	filename = replaceField(filename, FieldDate, movie.ReleaseDate)
	filename = replaceField(filename, FieldExt, fileMovie.Extension)
	return filename
}

func GenerateEpisodeFilename(pattern string, show mediadata.TvShow, episode mediadata.Episode, fileEpisode mediascanner.Episode) string {
	filename := pattern
	filename = replaceField(filename, FieldName, show.Title)
	filename = replaceField(filename, FieldYear, show.Year)
	filename = replaceFieldInt(filename, FieldSeason, episode.SeasonNumber)
	filename = replaceFieldInt(filename, FieldEpisode, episode.EpisodeNumber)
	filename = replaceField(filename, FieldEpisodeTitle, episode.Name)
	filename = replaceField(filename, FieldExt, fileEpisode.Extension)
	return filename
}

func generateDefaultMovieFilename(fileMovie mediascanner.Movie) string {
	filename := fileMovie.Name
	if fileMovie.Year != 0 {
		filename += " - " + strconv.Itoa(fileMovie.Year)
	}
	filename += fileMovie.Extension
	return filename
}

func replaceField(pattern string, field Field, value string) string {
	return strings.ReplaceAll(pattern, string(field), value)
}

func replaceFieldInt(pattern string, field Field, value int) string {
	return strings.ReplaceAll(pattern, string(field), fmt.Sprintf("%02d", value))
}
