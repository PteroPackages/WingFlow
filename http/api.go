package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	// allow override here
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req
}

func (c *Client) route(path string) string {
	path = strings.Replace(path, ":id", c.ID, 1)

	return fmt.Sprintf("%s/api/client%s", c.URL, path)
}

func (c *Client) Test() (bool, int, error) {
	req, _ := http.NewRequest("HEAD", c.route(""), nil)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return false, res.StatusCode, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusFound {
		return false, res.StatusCode, errors.New("recieved an invalid status from the api")
	}

	return true, res.StatusCode, nil
}

func (c *Client) GetUploadURL() (string, error) {
	body := bytes.Buffer{}
	req, _ := http.NewRequest("GET", c.route("/servers/:id/files/upload"), &body)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("received an invalid status from the api")
	}

	var data struct {
		Attributes struct {
			URL string
		}
	}

	defer res.Body.Close()
	buf, _ := io.ReadAll(res.Body)
	json.Unmarshal(buf, &data)

	return data.Attributes.URL, nil
}

func (c *Client) CreateDirectory(root, name string) error {
	type payload struct {
		Root string
		Name string
	}

	body := bytes.Buffer{}
	buf, _ := json.Marshal(payload{Root: root, Name: name})
	body.Write(buf)
	req, _ := http.NewRequest("POST", c.URL, &body)
	req = c.addHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	var data struct {
		Error string
	}
	defer res.Body.Close()
	buf, _ = io.ReadAll(res.Body)
	json.Unmarshal(buf, &data)

	return fmt.Errorf(data.Error)
}

func (c *Client) UploadFile(name string, file *os.File) error {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	part, _ := writer.CreateFormFile("files", name)
	io.Copy(part, file)
	writer.Close()

	url, err := c.GetUploadURL()
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", url, &body)
	req = c.addHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Mime-Type", "application/gzip")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	var data struct {
		Error string
	}
	defer res.Body.Close()
	buf, _ := io.ReadAll(res.Body)
	json.Unmarshal(buf, &data)

	return fmt.Errorf(data.Error)
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
