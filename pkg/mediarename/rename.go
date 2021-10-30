package mediarename

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kit/log"
)

const (
	apiBase = "https://api.tvmaze.com/"
)

var (
	extensions = map[string]bool{
		".mp4": true,
		".mkv": true,
	}
)

// TODO: Turn this into something less gross

func findFiles(base string) []string {
	var out []string

	filepath.Walk(base, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := path.Ext(p)
		if _, ok := extensions[ext]; ok {
			out = append(out, p)
		}
		return nil
	})

	return out
}

func generateName(base string, file string, show *Show, episode *Episode) string {
	newFile := fmt.Sprintf(
		"%s-s%0de%02d-%s%s",
		sanitize(show.Name),
		episode.Season,
		episode.Number,
		sanitize(episode.Name),
		path.Ext(file),
	)

	newPath := path.Join(
		base,
		sanitize(show.Name),
		fmt.Sprintf("season_%02d", episode.Season),
		newFile,
	)

	return newPath
}

func sanitize(val string) string {
	val = strings.Replace(val, " ", "_", -1)
	val = strings.ToLower(val)
	return val
}

func RenameMedia(src string, dest string, showId string, dryRun bool, logger log.Logger) error {
	files := findFiles(src)
	if len(files) == 0 {
		return fmt.Errorf("no files found to rename under %s", src)
	}

	client, err := NewTvMazeClient(apiBase, &http.Client{Timeout: 10 * time.Second}, logger)
	if err != nil {
		return err
	}

	show, err := client.ShowByImdb(showId)
	if err != nil {
		return err
	}

	episodes, err := client.Episodes(show)
	if err != nil {
		return err
	}

	lookup := NewEpisodeLookup(episodes, logger)
	for _, file := range files {
		f := path.Base(file)
		e, err := lookup.FindEpisode(f)
		if err != nil {
			return err
		}

		newPath := generateName(dest, f, show, e)
		fmt.Printf("NEW: %+v\n", newPath)
	}

	return nil
}
