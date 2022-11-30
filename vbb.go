package vbb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const BaseURL = "https://v5.vbb.transport.rest"

// Client is a VBB API client
type Client struct {
	endpoint string
	c        *http.Client
}

// New returns a new instance of VBB API client
func New(endpoint string, c *http.Client) *Client {
	if c == nil {
		c = http.DefaultClient
	}

	return &Client{
		endpoint: endpoint,
		c:        c,
	}
}

// Location is a station, stop, POI or an address
type Location struct {
	Type    string
	ID      string
	Name    string
	Address string
}

// LocationType represents location type
type LocationType uint8

const (
	LocationTypeUnknown LocationType = 0
	LocationTypeStop    LocationType = 1 << iota
	LocationTypeAddress
	LocationTypePOI
)

// Locations returns first resultsNum locations matching the query
func (c *Client) Locations(query string, locType LocationType, resultsNum int) ([]Location, error) {
	q := make(url.Values)
	q.Set("results", strconv.Itoa(resultsNum))
	q.Set("query", query)
	q.Set("pretty", "false")

	q.Set("stops", strconv.FormatBool(locType&LocationTypeStop != 0))
	q.Set("addresss", strconv.FormatBool(locType&LocationTypeAddress != 0))
	q.Set("poi", strconv.FormatBool(locType&LocationTypePOI != 0))

	data, err := c.sendRequest(http.MethodGet, "/locations?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve locations: %w", err)
	}

	defer data.Close()

	var res []Location
	if err := json.NewDecoder(data).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

func (c *Client) sendRequest(method, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, c.endpoint+url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %s: %w", url, err)
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", url, err)
	}

	return resp.Body, nil
}
