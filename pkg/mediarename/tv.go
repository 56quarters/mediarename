package mediarename

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type RenameType int

const (
	RenameCopy RenameType = iota
	RenameMove
)

type Rename struct {
	Old string
	New string
}

type TvRenamer struct {
	client MediaClient
	op     RenameType // ignored ATM
	commit bool
	logger log.Logger
}

func NewTvRenamer(client MediaClient, commit bool, logger log.Logger) *TvRenamer {
	return &TvRenamer{
		client: client,
		op:     RenameMove,
		commit: commit,
		logger: logger,
	}
}

func (r *TvRenamer) FindFiles(base string, extensions map[string]struct{}) ([]string, error) {
	var out []string

	err := filepath.Walk(base, func(p string, info fs.FileInfo, err error) error {
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

	if err != nil {
		return nil, fmt.Errorf("unable to find files: %w", err)
	}

	return out, nil
}

// TODO: Return a sorted list of a two element struct instead of a map for nicer output

func (r *TvRenamer) GenerateNames(files []string, dest string, imdb ImdbID) ([]Rename, error) {
	show, err := r.client.ShowByImdb(imdb)
	if err != nil {
		return nil, fmt.Errorf("show lookup error for imdb ID %s: %w", imdb, err)
	}

	episodes, err := r.client.Episodes(show)
	if err != nil {
		return nil, fmt.Errorf("episode lookup error show %s (%d): %w", show.Name, show.ID, err)
	}

	lookup := NewEpisodeLookup(episodes, r.logger)
	out := make([]Rename, 0, len(episodes))

	for _, file := range files {
		matched, err := lookup.FindEpisodes(file)
		if err != nil {
			level.Warn(r.logger).Log("msg", "unable to generate new name for file", "file", file, "err", err)
			continue
		}

		newName := r.nameFromEpisodes(file, dest, show, matched)
		out = append(out, Rename{
			Old: file,
			New: newName,
		})
	}

	return out, nil
}

func (r *TvRenamer) nameFromEpisodes(file string, dest string, show *Show, episodes Episodes) string {
	ext := path.Ext(file)

	// If there are multiple episodes that match for this particular file (such as when a
	// finale is two episodes but aired at the same time and thus the same file): use the first
	// match to generate the file name but append each episode after that with "-eXX" after
	// the primary tag.
	first := episodes[0]
	episodes = episodes[1:]
	tag := strings.Builder{}
	tag.WriteString(fmt.Sprintf("s%02de%02d", first.Season, first.Number))

	for _, e := range episodes {
		tag.WriteString(fmt.Sprintf("-e%02d", e.Number))
	}

	newFile := fmt.Sprintf(
		"%s-%s-%s%s",
		sanitize(show.Name),
		tag.String(),
		sanitize(first.Name),
		ext,
	)

	return path.Join(
		dest,
		sanitize(show.Name),
		fmt.Sprintf("season_%02d", first.Season),
		newFile,
	)
}

func (r *TvRenamer) RenameFiles(renames []Rename) error {
	for _, op := range renames {
		level.Info(r.logger).Log("old", op.Old, "new", op.New)

		if r.commit {
			dir := path.Dir(op.New)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("unable to create parent directory %s: %w", dir, err)
			}

			err = os.Rename(op.Old, op.New)
			if err != nil {
				return fmt.Errorf("unable to rename %s to %s: %w", op.Old, op.New, err)
			}
		}
	}

	return nil
}

func sanitize(val string) string {
	val = strings.Replace(val, " ", "_", -1)
	val = strings.Replace(val, ":", "", -1)
	val = strings.Replace(val, "'", "", -1)
	val = strings.Replace(val, "&", "and", -1)
	val = strings.ToLower(val)
	return val
}
