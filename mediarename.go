package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/56quarters/mediarename/pkg/mediarename"
)

const (
	apiBase = "https://api.tvmaze.com/"
)

var (
	extensions = map[string]struct{}{
		".avi": {},
		".idx": {},
		".mp4": {},
		".mkv": {},
		".sub": {},
	}
)

func setupLogger(l level.Option) log.Logger {
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	logger = level.NewFilter(logger, l)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	logger := setupLogger(level.AllowDebug())
	kp := kingpin.New(os.Args[0], "mediarename: rename media files based on their metadata")

	tv := kp.Command("tv", "rename TV episodes based on show and episode metadata ")
	tvID := tv.Arg("id", "IMDB show ID").Required().String()
	tvSrc := tv.Arg("src", "Directory of files to rename").Required().String()
	tvDest := tv.Arg("dest", "Destination of renamed files").Required().String()
	tvCommit := tv.Flag("commit", "Actually rename things instead of just printing new names.").Default("false").Bool()

	command, err := kp.Parse(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse CLI options", "err", err)
		return 1
	}

	switch command {
	case tv.FullCommand():
		if err := renameTv(*tvSrc, *tvDest, *tvID, *tvCommit, logger); err != nil {
			level.Error(logger).Log("msg", "failed to rename tv episodes", "err", err)
			return 1
		}
	}

	return 0
}

func renameTv(src string, dest string, showID string, commit bool, logger log.Logger) error {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	client, err := mediarename.NewTvMazeClient(apiBase, httpClient, logger)
	if err != nil {
		return err
	}

	renamer := mediarename.NewTvRenamer(client, commit, logger)
	files, err := renamer.FindFiles(src, extensions)
	if err != nil {
		return err
	}

	renames, err := renamer.GenerateNames(files, dest, mediarename.ImdbID(showID))
	if err != nil {
		return err
	}

	return renamer.RenameFiles(renames)
}
