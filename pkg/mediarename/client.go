package mediarename

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

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

