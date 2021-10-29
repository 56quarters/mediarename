package main

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/56quarters/mediarename/pkg/mediarename"
)

func setupLogger(l level.Option) log.Logger {
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	logger = level.NewFilter(logger, l)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}

func main() {
	logger := setupLogger(level.AllowDebug())

	kp := kingpin.New(os.Args[0], "mediarename: TBD")

	rename := kp.Command("rename", "rename things")
	renameId := rename.Arg("id", "IMDB show ID").Required().String()
	renameSrc := rename.Arg("src", "Files to rename").Required().String()
	renameDest := rename.Arg("dest", "Destination of renamed files").Required().String()
	renameDryRun := rename.Flag("dry-run", "Don't rename things.").Default("true").Bool()

	command, err := kp.Parse(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse CLI options", "err", err)
		os.Exit(1)
	}

	switch command {
	case rename.FullCommand():
		if err := mediarename.RenameMedia(*renameSrc, *renameDest, *renameId, *renameDryRun, logger); err != nil {
			level.Error(logger).Log("msg", "failed to lookup show", "err", err)
			os.Exit(1)
		}
	}
}
