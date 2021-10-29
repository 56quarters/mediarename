package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
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

var episodeRegex = regexp.MustCompile("(s[\\d]{2}e[\\d]{2})")

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
	renameDryRun := rename.Flag("dry-run", "Don't rename things.").Default("true").Bool()

	command, err := kp.Parse(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse CLI options", "err", err)
		os.Exit(1)
	}

	switch command {
	case rename.FullCommand():
		if err := renameMedia(*renameId, *renameDryRun, logger); err != nil {
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

	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup show by ID: %w", err)
	}

	defer func() { _ = res.Body.Close() }()
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
	path := fmt.Sprintf("shows/%d/episodes", show.Id)
	r := c.request(path, "")

	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup episodes by show ID %d: %w", show.Id, err)
	}

	defer func() { _ = res.Body.Close() }()
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
	p = strings.ToLower(p)
	file := path.Base(p)

	matched := episodeRegex.FindString(file)
	if matched == "" {
		return nil, fmt.Errorf("could not find season and episode in %s", file)
	}

	episode, ok := l.lookup[matched]
	if !ok {
		return nil, fmt.Errorf("not a known episode: %s", matched)
	}

	return &episode, nil
}

func renameMedia(showId string, dryRun bool, logger log.Logger) error {
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
	e, err := lookup.FindEpisode("something/blah-s01e01-blah.mp4")
	if err != nil {
		return err
	}

	fmt.Printf("RES: %+v\n", e)
	return nil
}
