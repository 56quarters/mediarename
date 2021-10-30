package mediarename

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func TestEpisodeLookup_FindEpisode(t *testing.T) {
	t.Run("no match in file name", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, log.NewNopLogger())
		e, err := lookup.FindEpisode("show-season_1_episode_1-pilot.mkv")

		assert.Nil(t, e)
		assert.Error(t, err)
	})

	t.Run("no episode available", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, log.NewNopLogger())
		e, err := lookup.FindEpisode("show-s01e03-something.mkv")

		assert.Nil(t, e)
		assert.Error(t, err)
	})

	t.Run("multi-episode match in file name", func(t *testing.T) {
		// TODO: Handle multi-episode case
		t.Skip()

		lookup := NewEpisodeLookup(testEpisodes, log.NewNopLogger())
		e, err := lookup.FindEpisode("show-s01e01-02-pilot.mkv")

		assert.Nil(t, e)
		assert.Error(t, err)
	})

	t.Run("success case", func(t *testing.T) {
		lookup := NewEpisodeLookup(testEpisodes, log.NewNopLogger())
		e, err := lookup.FindEpisode("show-s01e01-pilot.mkv")

		require.NoError(t, err)
		assert.Equal(t, &testEpisodes[0], e)
	})
}
