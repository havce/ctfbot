package ctftime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/havce/havcebot"
)

type Client struct {
	c *http.Client
}

func NewClient() *Client {
	return &Client{
		c: http.DefaultClient,
	}
}

func (c *Client) FindEventByID(ctx context.Context, id int) (*Event, error) {
	u, err := url.Parse("https://ctftime.org/api/v1/events/")
	if err != nil {
		return nil, err
	}

	u = u.JoinPath(strconv.Itoa(id), "/")

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Workaround for CTFtime API.
	req.Header.Add("User-Agent", "curl/8.5.0")

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 400 && resp.StatusCode < 499 {
		return nil, havcebot.Errorf(havcebot.ENOTFOUND, "Event not found.")
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, havcebot.Errorf(havcebot.EINVALID, "Internal server error.")
	}

	event := &Event{}
	if err := json.NewDecoder(resp.Body).Decode(event); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return event, err
}

func (c *Client) FindEvents(ctx context.Context, filter EventFilter) ([]*Event, error) {
	u, err := url.Parse("https://ctftime.org/api/v1/events/")
	if err != nil {
		return nil, err
	}

	q := u.Query()

	if filter.Limit != 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}

	if filter.Start != nil {
		q.Set("start", strconv.FormatInt(filter.Start.Unix(), 10))
	}

	if filter.Finish != nil {
		q.Set("finish", strconv.FormatInt(filter.Finish.Unix(), 10))
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Workaround for CTFtime API.
	req.Header.Add("User-Agent", "curl/8.5.0")

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	events := make([]*Event, 0)
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return events, err
}
