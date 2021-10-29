package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	apiBase = "https://api.tvmaze.com/"
)

var (
	episodeRegex = regexp.MustCompile("(?i)(s[\\d]{2}e[\\d]{2})")
	extensions   = map[string]bool{
		".mp4": true,
		".mkv": true,
	}
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
		if err := renameMedia(*renameSrc, *renameDest, *renameId, *renameDryRun, logger); err != nil {
			level.Error(logger).Log("msg", "failed to lookup show", "err", err)
			os.Exit(1)
		}
	}
}

type ErrorResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  int    `json:"status"`

	Previous struct {
		Name    string `json:"name"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"previous"`
}

type ShowExternals struct {
	TvRage  int    `json:"tvrage"`
	TheTvDb int    `json:"thetvdb"`
	Imdb    string `json:"imdb"`
}

type Show struct {
	Id        int           `json:"id"`
	Url       string        `json:"url"`
	Name      string        `json:"name"`
	Externals ShowExternals `json:"externals"`
}

type Episodes []Episode

type Episode struct {
	Id     int    `json:"id"`
	Url    string `json:"url"`
	Name   string `json:"name"`
	Season int    `json:"season"`
	Number int    `json:"number"`
	Type   string `json:"type"`
}

type TvMazeClient struct {
	client  *http.Client
	baseUrl *url.URL
	logger  log.Logger
}

func NewTvMazeClient(base string, client *http.Client, logger log.Logger) (*TvMazeClient, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base URL: %w", err)
	}

	return &TvMazeClient{
		client:  client,
		baseUrl: u,
		logger:  logger,
	}, nil
}

func (c *TvMazeClient) ShowByImdb(imdb string) (*Show, error) {
	params := url.Values{"imdb": {imdb}}
	r := c.request("lookup/shows", params.Encode())

	level.Debug(c.logger).Log("msg", "looking up show by imdb ID", "id", imdb, "url", r)
	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup show by ID: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	// TODO: Better error handling
	level.Debug(c.logger).Log("msg", "API response", "status", res.Status)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-success status code: %d", res.StatusCode)
	}

	var show Show
	err = json.NewDecoder(res.Body).Decode(&show)
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	return &show, nil
}

func (c *TvMazeClient) Episodes(show *Show) (Episodes, error) {
	p := fmt.Sprintf("shows/%d/episodes", show.Id)
	r := c.request(p, "")

	level.Debug(c.logger).Log("msg", "looking up episodes by native ID", "id", show.Id, "url", r)
	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup episodes by show ID %d: %w", show.Id, err)
	}

	defer func() { _ = res.Body.Close() }()

	// TODO: Better error handling
	level.Debug(c.logger).Log("msg", "API response", "status", res.Status)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-success status code %d", res.StatusCode)
	}

	var episodes Episodes
	err = json.NewDecoder(res.Body).Decode(&episodes)
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	return episodes, nil
}

func (c *TvMazeClient) request(path string, params string) string {
	requestUrl := url.URL{
		Scheme:   c.baseUrl.Scheme,
		Opaque:   c.baseUrl.Opaque,
		User:     c.baseUrl.User,
		Host:     c.baseUrl.Host,
		Path:     path,
		RawQuery: params,
	}

	return requestUrl.String()
}

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
		fmt.Sprintf("season_%02d", episode.Number),
		newFile,
	)

	return newPath
}

func sanitize(val string) string {
	val = strings.Replace(val, " ", "_", -1)
	val = strings.ToLower(val)
	return val
}

func renameMedia(src string, dest string, showId string, dryRun bool, logger log.Logger) error {
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
