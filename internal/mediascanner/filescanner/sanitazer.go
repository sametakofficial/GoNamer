package filescanner

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/config"
	"github.com/nouuu/gonamer/pkg/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	capitaliser = cases.Title(language.Turkish)
	defaultDeleteRegex *regexp.Regexp
	spaceRegex         = regexp.MustCompile(`[\._]`) // Bu basit olduğu için kalabilir.
	extractDateRegex   = regexp.MustCompile(`^(.+?)\s?\(?(19\d{2}|20\d{2})\)?.*$`)
	episodePatterns    []*regexp.Regexp
)

func init() {
	qualityKeywords := []string{"4K", "2160p", "1080p", "720p", "480p", "UHD", "BluRay", "BDRip", "BRRip", "WEB-DL", "WEBDL", "WEBRip", "HDTV", "DVDRip", "PROPER", "EXTENDED"}
	codecKeywords := []string{"x264", "x265", `H\.?264`, `H\.?265`, "HEVC", "10bit", "DTS-HD", "DTS", "Atmos", "TrueHD", "AC3", `DD\+?5\.1`, "AAC"}
	languageKeywords := []string{"FR EN", "MULTI", "TRUEFRENCH", "FRENCH", "VFF", "DUAL", "TR-EN", `TR\s?Dublaj`, "Dublaj"}

	allDefaultKeywords := append(qualityKeywords, codecKeywords...)
	allDefaultKeywords = append(allDefaultKeywords, languageKeywords...)

	joinedDefaultKeywords := strings.Join(allDefaultKeywords, "|")

	deleteRegexPattern := fmt.Sprintf(`(?i)[\s\._]?(%s)\b|[\[\(].*?[\]\)]|\s+$`, joinedDefaultKeywords)
	defaultDeleteRegex = regexp.MustCompile(deleteRegexPattern)

	episodeKeywords := []string{
		"episode",     // en
		"集",          // zh
		"एपिसोड",     // hi
		"episodio",    // es
		"épisode",     // fr
		"حلقة",        // ar
		"এপিসোড",     // bn
		"эпизод",      // ru
		"episódio",    // pt
		"episode",     // id
		"قسط",         // ur
		"Folge",       // de
		"エピソード",   // ja
		"kipindi",     // sw
		"प्रकरण",     // mr
		"பகுதி",       // ta
		"ఎపిసోడ్",    // te
		"tập",         // vi
		"에피소드",    // ko
		"bölüm",       // tr
		"episodio",    // it
		"odcinek",     // pl
		"ਭਾਗ",        // pa
		"episode",     // jv
		"એપિસોડ",    // gu
		"എപിസോഡ്",   // ml
		"एपिसोड",     // ne
		"කඩුව",       // si
		"ตอน",        // th
		"အပိုင်း",     // my
	}
	joinedEpisodeKeywords := strings.Join(episodeKeywords, "|")

	episodePatterns = []*regexp.Regexp{
		regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<name>.+?)[\. ]S(?P<season>\d{1,2})E(?P<episode>\d{1,3})`)),
		regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<name>.+?)[\. ](?P<season>\d{1,2})x(?P<episode>\d{1,3})`)),
		regexp.MustCompile(fmt.Sprintf(`(?i)^(?P<name>.+?)[\. ](%s)[\. ](?P<episode>\d{1,3})`, joinedEpisodeKeywords)),
	}
}
func parseMovieFileName(ctx context.Context, fileName string, cfg *config.Config) (movie mediascanner.Movie) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	movie.OriginalFilename = filename
	movie.FullPath = fileName
	movie.Extension = ext

	movie.Name, movie.Year = sanitizeMovieName(ctx, nameWithoutExt, cfg)

	return
}

func parseEpisodeFileName(ctx context.Context, fileName string, cfg *config.Config) (episode mediascanner.Episode) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	episode.OriginalFilename = filename
	episode.FullPath = fileName
	episode.Extension = ext

	var ignore bool
	episode.Name, episode.Season, episode.Episode, ignore = sanitizeEpisodeName(ctx, nameWithoutExt, cfg)

	if ignore {
		episode = mediascanner.Episode{
			OriginalFilename: filename,
		}
	}
	return
}

func sanitizeMovieName(ctx context.Context, nameWithoutExt string, cfg *config.Config) (name string, year int) {
	log := logger.FromContext(ctx)
	nameWithoutExt = sanitizeString(nameWithoutExt, cfg)

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

func sanitizeEpisodeName(ctx context.Context, nameWithoutExt string, cfg *config.Config) (name string, season int, episode int, ignore bool) {
	cleanedName := sanitizeString(nameWithoutExt, cfg)

	return parseEpisodeName(ctx, cleanedName, cfg.Scanner.ExcludeUnparsed)
}


func parseEpisodeName(ctx context.Context, name string, excludeUnparsed bool) (string, int, int, bool) {
	log := logger.FromContext(ctx)
	for _, pattern := range episodePatterns {
		matches := pattern.FindStringSubmatch(name)
		if len(matches) > 0 {
			result := make(map[string]string)
			for i, groupName := range pattern.SubexpNames() {
				if i != 0 && groupName != "" {
					result[groupName] = matches[i]
				}
			}
			showName := strings.TrimSpace(result["name"])
			seasonStr, ok := result["season"]
			if !ok {
				seasonStr = "1"
			}
			season, _ := strconv.Atoi(seasonStr)
			episode, _ := strconv.Atoi(result["episode"])
			return showName, season, episode, false
		}
	}
	log.With("name", name).Debug("No episode pattern matched.")
	if excludeUnparsed {
		return name, 0, 0, true
	}
	return name, 1, 1, false
}


func sanitizeString(str string, cfg *config.Config) string {
	str = spaceRegex.ReplaceAllString(str, " ")

	str = defaultDeleteRegex.ReplaceAllString(str, "")

	if cfg != nil && len(cfg.Scanner.DeleteKeywords) > 0 {
		for _, keyword := range cfg.Scanner.DeleteKeywords {
			if keyword == "" {
				continue
			}
			customRegex := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
			str = customRegex.ReplaceAllString(str, "")
		}
	}
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return capitaliser.String(strings.TrimSpace(str))
}

