package mediarename

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

const userAgent = "mediarename/0.1.0 (https://github.com/56quarters/mediarename)"

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

type Show struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Name      string `json:"name"`
	Externals struct {
		TvRage  int    `json:"tvrage"`
		TheTvDb int    `json:"thetvdb"`
		Imdb    string `json:"imdb"`
	} `json:"externals"`
}

type Episodes []Episode

type Episode struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Name   string `json:"name"`
	Season int    `json:"season"`
	Number int    `json:"number"`
	Type   string `json:"type"`
}

type ImdbID string

type MediaClient interface {
	ShowByImdb(imdb ImdbID) (*Show, error)
	Episodes(show *Show) (Episodes, error)
}

type TvMazeClient struct {
	client  *http.Client
	baseURL *url.URL
	logger  *slog.Logger
}

func NewTvMazeClient(base string, client *http.Client, logger *slog.Logger) (*TvMazeClient, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base URL: %w", err)
	}

	return &TvMazeClient{
		client:  client,
		baseURL: u,
		logger:  logger,
	}, nil
}

// ShowByImdb implements the MediaClient interface
func (c *TvMazeClient) ShowByImdb(imdb ImdbID) (*Show, error) {
	params := url.Values{"imdb": {string(imdb)}}
	r, err := c.request("lookup/shows", params.Encode())
	if err != nil {
		return nil, fmt.Errorf("unable to build request for show by ID %s: %w", imdb, err)
	}

	c.logger.Debug("looking up show by imdb ID", "id", imdb, "url", r.URL)
	res, err := c.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup show by ID: %w", err)
	}

	defer c.drainAndClose(res.Body)

	// TODO: Better error handling
	c.logger.Debug("API response", "status", res.Status)
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

// Episodes implements the MediaClient interface
func (c *TvMazeClient) Episodes(show *Show) (Episodes, error) {
	p := fmt.Sprintf("shows/%d/episodes", show.ID)
	r, err := c.request(p, "")
	if err != nil {
		return nil, fmt.Errorf("unable to build request for episodes by native ID %d: %w", show.ID, err)
	}

	c.logger.Debug("looking up episodes by native ID", "id", show.ID, "url", r.URL)
	res, err := c.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup episodes by show ID %d: %w", show.ID, err)
	}

	defer c.drainAndClose(res.Body)

	// TODO: Better error handling
	c.logger.Debug("API response", "status", res.Status)
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

func (c *TvMazeClient) request(path string, params string) (*http.Request, error) {
	requestURL := url.URL{
		Scheme:   c.baseURL.Scheme,
		Opaque:   c.baseURL.Opaque,
		User:     c.baseURL.User,
		Host:     c.baseURL.Host,
		Path:     path,
		RawQuery: params,
	}

	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", userAgent)
	return req, nil
}

func (c *TvMazeClient) drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
