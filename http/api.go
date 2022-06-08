package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	URL    string
	Key    string
	ID     string
	client http.Client
}

func New(url, key, id string) *Client {
	return &Client{
		URL:    url,
		Key:    key,
		ID:     id,
		client: http.Client{},
	}
}

func (c *Client) addHeaders(req *http.Request) *http.Request {
	req.Header.Add("User-Agent", "WingFlow Client")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Key))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req
}

func (c *Client) route(path string) string {
	path = strings.Replace(path, ":id", c.ID, 1)

	return fmt.Sprintf("%s/api/client%s", c.URL, path)
}

func (c *Client) Test() (int, error) {
	req, _ := http.NewRequest("HEAD", c.route(""), nil)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusFound {
		return res.StatusCode, fmt.Errorf("could not reach the api (status: %d)", res.StatusCode)
	}

	return res.StatusCode, nil
}

func (c *Client) GetRootFiles() ([]string, error) {
	req, _ := http.NewRequest("GET", c.route("/servers/:id/files/list"), nil)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unknown error: %d", res.StatusCode)
	}

	type file struct {
		Name string `json:"name"`
	}
	var wrapper struct {
		Data []struct {
			Attributes file `json:"attributes"`
		} `json:"data"`
	}

	defer res.Body.Close()
	buf, _ := io.ReadAll(res.Body)
	json.Unmarshal(buf, &wrapper)

	var names []string
	for _, d := range wrapper.Data {
		names = append(names, d.Attributes.Name)
	}

	return names, nil
}

func (c *Client) WriteFile(path, name string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	body := bytes.Buffer{}
	body.Write(buf)

	req, _ := http.NewRequest("POST", c.route("/servers/:id/files/write?file="+name), &body)
	req = c.addHeaders(req)
	req.Header.Set("Content-Type", "text/plain")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	// TODO: implement fractal
	return fmt.Errorf("unknown error")
}

func (c *Client) DeleteFiles(paths []string) error {
	data := struct {
		Root  string   `json:"root"`
		Files []string `json:"files"`
	}{Root: "/", Files: paths}

	buf, _ := json.Marshal(data)
	body := bytes.Buffer{}
	body.Write(buf)

	req, _ := http.NewRequest("POST", c.route("/servers/:id/files/delete"), &body)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("unknown error: %d", res.StatusCode)
}

func (c *Client) SetPower(state string) error {
	data := struct {
		Signal string `json:"signal"`
	}{Signal: state}

	buf, _ := json.Marshal(data)
	body := bytes.Buffer{}
	body.Write(buf)

	req, _ := http.NewRequest("POST", c.route("/servers/:id/power"), &body)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("unknown error: %d", res.StatusCode)
}

func (c *Client) CompressFiles(paths []string) (string, error) {
	data := struct {
		Root  string   `json:"root"`
		Files []string `json:"files"`
	}{Root: "/", Files: paths}

	buf, _ := json.Marshal(data)
	body := bytes.Buffer{}
	body.Write(buf)

	req, _ := http.NewRequest("POST", c.route("/servers/:id/files/compress"), &body)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unknown error: %d", res.StatusCode)
	}

	var wrapper struct {
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	}

	defer res.Body.Close()
	buf, _ = io.ReadAll(res.Body)
	json.Unmarshal(buf, &wrapper)

	return wrapper.Attributes.Name, nil
}
