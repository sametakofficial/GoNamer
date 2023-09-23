package tmdb

import (
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
)

func TestCfgMap(t *testing.T) {
	Convey("Given a cfgMap function", t, func() {
		opt := AllOpts{
			APIKey: "123",
			Opts: Opts{
				Lang:  "fr",
				Adult: true,
			},
		}
		Convey("When it is called with opts", func() {
			newConf := cfgMap(opt)
			Convey("It should return the map with expected 'language' and 'include_adult' fields", func() {
				So(newConf["language"], ShouldEqual, "fr")
				So(newConf["include_adult"], ShouldEqual, strconv.FormatBool(true))
			})
		})
		Convey("When it is called with opts and override map", func() {
			override := map[string]string{
				"language":      "en",
				"include_adult": strconv.FormatBool(false),
			}
			newConf := cfgMap(opt, override)
			Convey("It should return the map with overridden 'language' and 'include_adult' fields", func() {
				So(newConf["language"], ShouldEqual, "en")
				So(newConf["include_adult"], ShouldEqual, strconv.FormatBool(false))
			})
		})
	})
}

func TestDefaultOpts(t *testing.T) {
	Convey("Given a defaultOpts function", t, func() {
		Convey("When it is called with API key", func() {
			opt := defaultOpts("123")
			Convey("It should return opts with default 'Lang' and 'Adult' fields and provided API key", func() {
				So(opt.APIKey, ShouldEqual, "123")
				So(opt.Lang, ShouldEqual, "en-US")
				So(opt.Adult, ShouldEqual, false)
			})
		})
	})
}
