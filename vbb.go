package vbb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	LocationTypeAny = LocationTypeStop | LocationTypeAddress | LocationTypePOI
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

// StopsNearby returns resultsNum stops within distance meters of walking from given location
func (c *Client) StopsNearby(lat, lng float64, distance, resultsNum int) ([]Location, error) {
	q := make(url.Values)
	q.Set("results", strconv.Itoa(resultsNum))
	q.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("pretty", "false")

	data, err := c.sendRequest(http.MethodGet, "/stops/nearby?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stops nearby: %w", err)
	}

	defer data.Close()

	var res []Location
	if err := json.NewDecoder(data).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

// TransportationType represents the type transport type
type TransportationType uint8

const (
	SuburbanTrain TransportationType = 1<<iota + 1
	Subway
	Tram
	Bus
	Ferry
	ExpressTrain
	RegionalTrain

	UrbanTransport = SuburbanTrain | Subway | Tram | Bus | Ferry | RegionalTrain
	AllTransport   = UrbanTransport | ExpressTrain
)

// Line represents a public transportation line
type Line struct {
	Name    string
	Product string
}

// Departure represents departure information
type Departure struct {
	Direction       string
	When            time.Time
	PlannedWhen     time.Time
	Delay           int
	Platform        int
	PlannedPlatform int
	Line            Line
}

// Departures returns a list of departures for the stop at given time
func (c *Client) Departures(stopID string, when time.Time, duration time.Duration, transportTypes TransportationType) ([]Departure, error) {
	q := make(url.Values)
	q.Set("when", when.Format(time.RFC3339))
	q.Set("duration", strconv.FormatFloat(duration.Minutes(), 'f', 0, 64))
	q.Set("pretty", "false")

	q.Set("suburban", strconv.FormatBool(transportTypes&SuburbanTrain != 0))
	q.Set("subway", strconv.FormatBool(transportTypes&Subway != 0))
	q.Set("tram", strconv.FormatBool(transportTypes&Tram != 0))
	q.Set("bus", strconv.FormatBool(transportTypes&Bus != 0))
	q.Set("ferry", strconv.FormatBool(transportTypes&Ferry != 0))
	q.Set("express", strconv.FormatBool(transportTypes&ExpressTrain != 0))
	q.Set("regional", strconv.FormatBool(transportTypes&RegionalTrain != 0))

	data, err := c.sendRequest(http.MethodGet, "/stops/"+stopID+"/departures")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve departures for %s: %w", stopID, err)
	}

	defer data.Close()

	var res []Departure
	if err := json.NewDecoder(data).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

// Arrivals returns a list of arrivals for the stop at given time
func (c *Client) Arrivals(stopID string, when time.Time, duration time.Duration, transportTypes TransportationType) ([]Departure, error) {
	q := make(url.Values)
	q.Set("when", when.Format(time.RFC3339))
	q.Set("duration", strconv.FormatFloat(duration.Minutes(), 'f', 0, 64))
	q.Set("pretty", "false")

	q.Set("suburban", strconv.FormatBool(transportTypes&SuburbanTrain != 0))
	q.Set("subway", strconv.FormatBool(transportTypes&Subway != 0))
	q.Set("tram", strconv.FormatBool(transportTypes&Tram != 0))
	q.Set("bus", strconv.FormatBool(transportTypes&Bus != 0))
	q.Set("ferry", strconv.FormatBool(transportTypes&Ferry != 0))
	q.Set("express", strconv.FormatBool(transportTypes&ExpressTrain != 0))
	q.Set("regional", strconv.FormatBool(transportTypes&RegionalTrain != 0))

	data, err := c.sendRequest(http.MethodGet, "/stops/"+stopID+"/arrivals")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve arrivals for %s: %w", stopID, err)
	}

	defer data.Close()

	var res []Departure
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
