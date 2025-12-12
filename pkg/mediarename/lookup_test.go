package mediarename

import (
	"log/slog"
	"testing"
)

var testEpisodes = Episodes{
	Episode{
		ID:     1,
		URL:    "https://api.example.com/show/1/episode/1",
		Name:   "Pilot",
		Season: 1,
		Number: 1,
		Type:   "regular",
	},
	Episode{
		ID:     2,
		URL:    "https://api.example.com/show/1/episode/2",
		Name:   "Events",
		Season: 1,
		Number: 2,
		Type:   "regular",
	},
	Episode{
		ID:     3,
		URL:    "https://api.example.com/show/1/episode/3",
		Name:   "Finale",
		Season: 1,
		Number: 123,
		Type:   "regular",
	},
}

func TestEpisodeLookup_FindEpisode(t *testing.T) {
	t.Run("no match in file name", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-season_1_episode_1-pilot.mkv")

		RequireEqual(t, 0, len(episodes))
		RequireErrorIs(t, err, ErrBadMetadata)
	})

	t.Run("no episode available", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-s01e03-something.mkv")

		RequireEqual(t, 0, len(episodes))
		RequireErrorIs(t, err, ErrUnknownEpisode)
	})

	t.Run("multi episode match in file name lowercase", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-s01e01-e02-pilot.mkv")

		RequireNoError(t, err)
		RequireEqual(t, 2, len(episodes))
		RequireEqual(t, testEpisodes[0], episodes[0])
		RequireEqual(t, testEpisodes[1], episodes[1])
	})

	t.Run("multi episode match in file name uppercase", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-S01E01-E02-pilot.mkv")

		RequireNoError(t, err)
		RequireEqual(t, 2, len(episodes))
		RequireEqual(t, testEpisodes[0], episodes[0])
		RequireEqual(t, testEpisodes[1], episodes[1])
	})

	t.Run("single episode match in file name", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-s01e01-pilot.mkv")

		RequireNoError(t, err)
		RequireEqual(t, 1, len(episodes))
		RequireEqual(t, testEpisodes[0], episodes[0])
	})

	t.Run("many digit episode number", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, slog.New(slog.DiscardHandler))
		episodes, err := lookup.FindEpisodes("show-s01e123-finale.mkv")

		RequireNoError(t, err)
		RequireEqual(t, 1, len(episodes))
		RequireEqual(t, testEpisodes[2], episodes[0])
	})
}
