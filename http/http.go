package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	http.Client
	URL string
	Key string
	ID  string
}

func New(url, key, id string) *Client {
	return &Client{http.Client{}, url, key, id}
}

func (c *Client) newRequest(method, route string, body io.Reader) *http.Request {
	r, _ := http.NewRequest(method, c.URL+route, body)
	r.Header.Set("User-Agent", "WingFlow Client")
	r.Header.Set("Authorizarion", fmt.Sprintf("Bearer %s", c.Key))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")

	return r
}

func (c *Client) TestConnection() (int, error) {
	req := c.newRequest("HEAD", fmt.Sprintf("/api/client/servers/%s", c.ID), nil)
	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, errors.New("connection to panel failed")
	}

	return 0, nil
}

func (c *Client) GetFiles() ([]string, error) {
	req := c.newRequest("GET", fmt.Sprintf("/api/client/servers/%s/files/list", c.ID), nil)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []struct {
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}

	defer res.Body.Close()
	buf, _ := io.ReadAll(res.Body)
	if err = json.Unmarshal(buf, &wrapper); err != nil {
		return nil, err
	}

	n := make([]string, len(wrapper.Data))
	for _, d := range wrapper.Data {
		n = append(n, d.Attributes.Name)
	}

	return n, nil
}

func (c *Client) UploadFile(w *io.Reader) error {
	return nil
}

func (c *Client) DeleteFiles(files []string) error {
	data, _ := json.Marshal(struct {
		Root  string   `json:"root"`
		Files []string `json:"files"`
	}{"/", files})
	buf := bytes.Buffer{}
	buf.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/api/client/servers/%s/files/delete", c.ID), &buf)
	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.New("failed to delete server files")
	}

	return nil
}

func (c *Client) SetPower(state string) error {
	data, _ := json.Marshal(struct {
		Signal string `json:"signal"`
	}{state})
	buf := bytes.Buffer{}
	buf.Write(data)

	req := c.newRequest("POST", fmt.Sprintf("/api/client/servers/%s/power", c.ID), &buf)
	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.New("failed to delete server files")
	}

	return nil
}

func (c *Client) CompressFiles(files []string) (string, error) {
	return "", nil
}
