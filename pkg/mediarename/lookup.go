package mediarename

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var (
	episodeRegex = regexp.MustCompile("(?i)(s[\\d]{2}e[\\d]{2})")
	multiRegex = regexp.MustCompile("(?i)(s[\\d]{2}e[\\d]{2}-[\\d]{2})")
)

type EpisodeLookup struct {
	lookup map[string]Episode
	logger log.Logger
}

func NewEpisodeLookup(episodes Episodes, logger log.Logger) *EpisodeLookup {
	lookup := make(map[string]Episode)

	for _, e := range episodes {
		lookup[fmt.Sprintf("s%02de%02d", e.Season, e.Number)] = e
	}

	return &EpisodeLookup{lookup: lookup, logger: logger}
}

func (l *EpisodeLookup) FindEpisode(p string) (*Episode, error) {
	file := path.Base(p)

	level.Debug(l.logger).Log("msg", "extracting season episode from file", "file", file)
	matched := episodeRegex.FindString(file)
	if matched == "" {
		return nil, fmt.Errorf("could not find season and episode in %s", file)
	}

	// regex is case-insensitive, but we store the lowercase version in the map
	matched = strings.ToLower(matched)

	level.Debug(l.logger).Log("msg", "using parsed season episode for lookup", "matched", matched)
	episode, ok := l.lookup[matched]
	if !ok {
		return nil, fmt.Errorf("not a known episode: %s", matched)
	}

	return &episode, nil
}
