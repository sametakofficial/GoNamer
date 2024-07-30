package filescanner

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nouuu/mediatracker/internal/mediascanner"
	"github.com/nouuu/mediatracker/pkg/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	capitaliser   = cases.Title(language.French)
	deleteRegexes = []*regexp.Regexp{
		regexp.MustCompile(`[\[\(].*?[\]\)]|-\s*\d+p.*`),
		regexp.MustCompile(`\s+$`),
		regexp.MustCompile(`\s(FR EN|FR-EN|MULTI|TRUEFRENCH|FRENCH|VFF)\s?`),
	}
	spaceRegexes = []*regexp.Regexp{
		regexp.MustCompile(`[^\pL\s_\d]+`),
	}
	extractDateRegex    = regexp.MustCompile(`^(.+)(19\d{2}|20\d{2}).*$`)
	tvShowRegex         = regexp.MustCompile(`^(.+?)[sS]?(\d+)[eExX](\d+)(?:.*|$)`)
	fallbackTvShowRegex = regexp.MustCompile(`^(.+?)(\d+)[a-z]{2}\s*Season\s*(\d+)(?:.*|$)`)
	episodeOnlyRegex    = regexp.MustCompile(`^(.+?)(\d{2,})(?:.*|$)`)
)

func parseMovieFileName(ctx context.Context, fileName string) (movie mediascanner.Movie) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	movie.OriginalFilename = filename
	movie.FullPath = fileName
	movie.Extension = ext

	movie.Name, movie.Year = sanitizeMovieName(ctx, nameWithoutExt)

	return
}

func parseEpisodeFileName(ctx context.Context, fileName string, excludeUnparsed bool) (episode mediascanner.Episode) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	episode.OriginalFilename = filename
	episode.FullPath = fileName
	episode.Extension = ext

	var ignore bool
	episode.Name, episode.Season, episode.Episode, ignore = sanitizeEpisodeName(ctx, nameWithoutExt, excludeUnparsed)

	if ignore {
		episode = mediascanner.Episode{
			OriginalFilename: filename,
		}
	}
	return
}

func sanitizeMovieName(ctx context.Context, nameWithoutExt string) (name string, year int) {
	log := logger.FromContext(ctx)
	nameWithoutExt = sanitizeString(nameWithoutExt)

	matches := extractDateRegex.FindStringSubmatch(nameWithoutExt)
	if len(matches) == 3 {
		name = strings.TrimSpace(matches[1])
		year, _ = strconv.Atoi(matches[2])
	} else {
		log.With("name", nameWithoutExt).Debug("Could not extract year from movie name")
		name = nameWithoutExt
	}
	return
}

func sanitizeEpisodeName(ctx context.Context, nameWithoutExt string, excludeUnparsed bool) (name string, season int, episode int, ignore bool) {
	nameWithoutExt = sanitizeString(nameWithoutExt)

	name, season, episode, ignore = parseEpisodeName(ctx, nameWithoutExt, excludeUnparsed)
	return
}

func parseEpisodeName(ctx context.Context, name string, excludeUnparsed bool) (string, int, int, bool) {
	log := logger.FromContext(ctx)

	if matches := tvShowRegex.FindStringSubmatch(name); len(matches) == 4 {
		return extractNameAndEpisode(matches, 1, 2, 3)
	}
	log.With("name", name).Debug("Could not extract season and episode from episode name, trying fallback regex")

	if matches := fallbackTvShowRegex.FindStringSubmatch(name); len(matches) == 4 {
		return extractNameAndEpisode(matches, 1, 2, 3)
	}
	log.With("name", name).Debug("Could not extract season and episode from episode name")

	if matches := episodeOnlyRegex.FindStringSubmatch(name); len(matches) == 3 {
		return extractNameAndEpisode(matches, 1, 0, 2) // Season set to 1
	}
	log.With("name", name).Debug("Could not extract episode from episode name")

	if excludeUnparsed {
		return name, 0, 0, true
	}
	return name, 1, 1, false
}

func extractNameAndEpisode(matches []string, nameIndex, seasonIndex, episodeIndex int) (string, int, int, bool) {
	name := strings.TrimSpace(matches[nameIndex])
	season := 1
	if seasonIndex != 0 {
		season, _ = strconv.Atoi(matches[seasonIndex])
	}
	episode, _ := strconv.Atoi(matches[episodeIndex])
	return name, season, episode, false
}

func sanitizeString(str string) string {
	for _, regex := range spaceRegexes {
		str = regex.ReplaceAllString(str, " ")
	}

	for _, regex := range deleteRegexes {
		str = regex.ReplaceAllString(str, "")
	}

	return capitaliser.String(strings.TrimSpace(str))
}
