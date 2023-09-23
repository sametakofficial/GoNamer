package tmdb

import (
	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"strconv"
)

const tmdbImageBaseUrl = "https://image.tmdb.org/t/p/original"

type OptFunc func(opts *Opts)

type AllOpts struct {
	APIKey string
	Opts
}

type Opts struct {
	Lang  string
	Adult bool
}

func WithLang(lang string) OptFunc {
	return func(opts *Opts) {
		opts.Lang = lang
	}
}

func WithAdult(adult bool) OptFunc {
	return func(opts *Opts) {
		opts.Adult = adult
	}
}

func defaultOpts(apiKey string) AllOpts {
	return AllOpts{
		APIKey: apiKey,
		Opts: Opts{
			Lang:  "en-US",
			Adult: false,
		},
	}
}

type tmdbClient struct {
	client *tmdb.Client
	opts   AllOpts
}

func cfgMap(opts AllOpts, args ...map[string]string) map[string]string {
	cfg := map[string]string{
		"language":      opts.Lang,
		"include_adult": strconv.FormatBool(opts.Adult),
	}

	for _, arg := range args {
		for k, v := range arg {
			cfg[k] = v
		}
	}
	return cfg
}

func buildGenres(genres []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) []mediadata.Genre {
	var g = make([]mediadata.Genre, len(genres))
	for i, genre := range genres {
		g[i] = mediadata.Genre{
			ID:   strconv.FormatInt(genre.ID, 10),
			Name: genre.Name,
		}
	}
	return g
}

func buildCast(cast []struct {
	Adult              bool    `json:"adult"`
	CastID             int64   `json:"cast_id"`
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Gender             int     `json:"gender"`
	ID                 int64   `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Order              int     `json:"order"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}) []mediadata.Person {
	var c = make([]mediadata.Person, len(cast))
	for i, person := range cast {
		c[i] = mediadata.Person{
			ID:         strconv.FormatInt(person.ID, 10),
			Name:       person.Name,
			Character:  person.Character,
			ProfileURL: tmdbImageBaseUrl + person.ProfilePath,
		}
	}
	return c
}

func buildStudio(studios []struct {
	Name          string `json:"name"`
	ID            int64  `json:"id"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}) []mediadata.Studio {
	var s = make([]mediadata.Studio, len(studios))
	for i, studio := range studios {
		s[i] = mediadata.Studio{
			ID:   strconv.FormatInt(studio.ID, 10),
			Name: studio.Name,
		}
	}
	return s
}
