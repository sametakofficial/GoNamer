package mediarenamer

import (
	"strconv"
	"strings"

	"github.com/nouuu/mediatracker/internal/mediadata"
	"github.com/nouuu/mediatracker/internal/mediascanner"
)

type Field string

const (
	FieldName Field = "{name}"
	FieldYear Field = "{year}"
	FieldDate Field = "{date}"
	FieldExt  Field = "{extension}"
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
