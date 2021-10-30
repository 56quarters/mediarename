package mediarename

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var (
	ErrBadMetadata    = errors.New("bad season or episode metadata")
	ErrUnknownEpisode = errors.New("unknown episode")

	multiRegex = regexp.MustCompile(`(?i)(s[\d]{2})(e[\d]{2})-?(e[\d]{2})?`)
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

func (l *EpisodeLookup) FindEpisodes(p string) (Episodes, error) {
	file := path.Base(p)

	level.Debug(l.logger).Log("msg", "extracting season episode from file", "file", file)
	matched := multiRegex.FindStringSubmatch(file)
	if matched == nil {
		return nil, fmt.Errorf("%w: could not find season and episode in %s", ErrBadMetadata, file)
	}

	// Note that we lowercase everything here since the episode tags are stored as lowercase
	var lookup []string
	if len(matched) == 4 {
		if matched[3] == "" {
			// Last match is empty, this must be the single episode case
			lookup = append(lookup, strings.ToLower(matched[1]+matched[2]))
		} else {
			// last match has something in it, multiple episode case
			lookup = append(lookup, strings.ToLower(matched[1]+matched[2]))
			lookup = append(lookup, strings.ToLower(matched[1]+matched[3]))
		}
	}

	var out []Episode
	for _, meta := range lookup {
		level.Debug(l.logger).Log("msg", "using parsed season episode for lookup", "meta", meta)
		e, ok := l.lookup[meta]
		if !ok {
			return nil, fmt.Errorf("%w: trying to match %s from %s", ErrUnknownEpisode, meta, file)
		}

		out = append(out, e)
	}

	return out, nil
}
