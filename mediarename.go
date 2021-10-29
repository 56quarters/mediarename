package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	apiBase    = "https://api.tvmaze.com/"
	showLookup = "lookup/shows"
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
	renameDryRun := rename.Flag("dry-run", "Don't rename things.").Default("true").Bool()

	command, err := kp.Parse(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse CLI options", "err", err)
		os.Exit(1)
	}

	switch command {
	case rename.FullCommand():
		if err := renameMedia(*renameId, *renameDryRun); err != nil {
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

type ShowResponse struct {
	Id        int           `json:"id"`
	Url       string        `json:"url"`
	Name      string        `json:"name"`
	Externals ShowExternals `json:"externals"`
}

type EpisodeResponse []Episode

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
}

func NewTvMazeClient(base string, client *http.Client) (*TvMazeClient, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base URL: %w", err)
	}

	return &TvMazeClient{
		client:  client,
		baseUrl: u,
	}, nil
}

func (c *TvMazeClient) ShowByImdb(imdb string) (*ShowResponse, error) {
	params := url.Values{"imdb": {imdb}}
	r := c.request("lookup/shows", params.Encode())

	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup show by ID: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-success status code: %d", res.StatusCode)
	}

	var successResponse ShowResponse
	err = json.NewDecoder(res.Body).Decode(&successResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	return &successResponse, nil
}

func (c *TvMazeClient) Episodes(show *ShowResponse) (*EpisodeResponse, error) {
	path := fmt.Sprintf("shows/%d/episodes", show.Id)
	r := c.request(path, "")

	res, err := c.client.Get(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup episodes by show ID %d: %w", show.Id, err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-success status code %d", res.StatusCode)
	}

	var successResponse EpisodeResponse
	err = json.NewDecoder(res.Body).Decode(&successResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	return &successResponse, nil
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

func renameMedia(showId string, dryRun bool) error {
	client, err := NewTvMazeClient(apiBase, &http.Client{})
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

	fmt.Printf("RES1: %+v\n", show)
	fmt.Printf("RES2: %+v\n", episodes)
	return nil
}
