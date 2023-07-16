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

const BaseURL = "https://v6.vbb.transport.rest"

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
	Type                string
	ID                  string
	Name                string
	Address             string
	Latitude, Longitude float64
	Distance            int
}

type hafasLocation struct {
	Type     string
	ID       string
	Name     string
	Address  string `json:",omitempty"`
	Location *struct {
		Type                string
		Latitude, Longitude float64
	} `json:",omitempty"`
	Latitude, Longitude float64 `json:",omitempty"`
	POI                 bool    `json:",omitempty"`
}

// MarshalJSON marshals Location into HAFAS@v6 JSON representation
func (loc Location) MarshalJSON() ([]byte, error) {
	hLoc := hafasLocation{
		Type:    loc.Type,
		ID:      loc.ID,
		Name:    loc.Name,
		Address: loc.Address,
		POI:     loc.Type == "poi",
	}

	if loc.Type == "stop" {
		hLoc.Location = &struct {
			Type                string
			Latitude, Longitude float64
		}{
			Type:      "location",
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		}
	}

	return json.Marshal(hLoc)
}

// UnmarshalJSON unmarshals HAFAS@v6 JSON representation into Location
func (loc *Location) UnmarshalJSON(data []byte) error {
	var hLoc hafasLocation

	if err := json.Unmarshal(data, &hLoc); err != nil {
		return fmt.Errorf("failed to unmarshal location: %w", err)
	}

	*loc = Location{
		Type:      hLoc.Type,
		ID:        hLoc.ID,
		Name:      hLoc.Name,
		Address:   hLoc.Address,
		Latitude:  hLoc.Latitude,
		Longitude: hLoc.Longitude,
	}

	if hLoc.Location != nil {
		loc.Latitude = hLoc.Location.Latitude
		loc.Longitude = hLoc.Location.Longitude
	}

	return nil
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
	q.Set("fuzzy", "false")

	q.Set("stops", strconv.FormatBool(locType&LocationTypeStop != 0))
	q.Set("addresss", strconv.FormatBool(locType&LocationTypeAddress != 0))
	q.Set("poi", strconv.FormatBool(locType&LocationTypePOI != 0))

	data, err := c.sendRequest(http.MethodGet, "/locations?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve locations: %w", err)
	}

	defer data.Close()

	var results []Location
	if err := json.NewDecoder(data).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// StopsNearby returns resultsNum stops within distance meters of walking from given location
func (c *Client) StopsNearby(lat, lng float64, distance, resultsNum int) ([]Location, error) {
	q := make(url.Values)
	q.Set("results", strconv.Itoa(resultsNum))
	q.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("distance", strconv.Itoa(distance))
	q.Set("pretty", "false")

	data, err := c.sendRequest(http.MethodGet, "/locations/nearby?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stops nearby: %w", err)
	}

	defer data.Close()

	var results []Location
	if err := json.NewDecoder(data).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
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
	Platform        string
	PlannedPlatform string
	Line            Line
}

// Departures returns a list of departures for the stop at given time
func (c *Client) Departures(stopID string, when time.Time, duration time.Duration, transportTypes TransportationType) ([]Departure, error) {
	q := addTransportTypeParams(make(url.Values), transportTypes)

	q.Set("when", when.Format("2006-01-02T15:04:05-0700"))
	q.Set("duration", strconv.Itoa(int(duration.Minutes())))
	q.Set("pretty", "false")

	data, err := c.sendRequest(http.MethodGet, "/stops/"+stopID+"/departures?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve departures for %s: %w", stopID, err)
	}

	defer data.Close()

	var res struct {
		Departures []Departure
	}
	if err := json.NewDecoder(data).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Departures, nil
}

// Arrivals returns a list of arrivals for the stop at given time
func (c *Client) Arrivals(stopID string, when time.Time, duration time.Duration, transportTypes TransportationType) ([]Departure, error) {
	q := addTransportTypeParams(make(url.Values), transportTypes)

	q.Set("when", when.Format("2006-01-02T15:04:05-0700"))
	q.Set("duration", strconv.Itoa(int(duration.Minutes())))
	q.Set("pretty", "false")

	data, err := c.sendRequest(http.MethodGet, "/stops/"+stopID+"/arrivals?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve arrivals for %s: %w", stopID, err)
	}

	defer data.Close()

	var res struct {
		Arrivals []struct {
			Departure
			Provenance string
		}
	}
	if err := json.NewDecoder(data).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var arrivals []Departure
	for _, a := range res.Arrivals {
		if a.Direction == "" {
			a.Direction = a.Provenance // VBB puts the direction in the provenance field for arrivals
		}
		arrivals = append(arrivals, a.Departure)
	}

	return arrivals, nil
}

func addTransportTypeParams(q url.Values, transportTypes TransportationType) url.Values {
	q.Set("suburban", strconv.FormatBool(transportTypes&SuburbanTrain != 0))
	q.Set("subway", strconv.FormatBool(transportTypes&Subway != 0))
	q.Set("tram", strconv.FormatBool(transportTypes&Tram != 0))
	q.Set("bus", strconv.FormatBool(transportTypes&Bus != 0))
	q.Set("ferry", strconv.FormatBool(transportTypes&Ferry != 0))
	q.Set("express", strconv.FormatBool(transportTypes&ExpressTrain != 0))
	q.Set("regional", strconv.FormatBool(transportTypes&RegionalTrain != 0))

	return q
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
