package tmdb

import (
	"github.com/cyruzin/golang-tmdb"
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
