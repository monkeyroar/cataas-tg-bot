package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const pathCat = "cat"
const pathSays = "says/%s"
const defaultTimeout = 5 * time.Second

type CatJson struct {
	Id        string   `json:"_id"`
	CreatedAt string   `json:"created_at"`
	Tags      []string `json:"tags"`
	Url       string   `json:"url"`
	MimeType  string   `json:"mimetype"`
}

type cataasAPI struct {
	baseUrl *url.URL
	client  *http.Client
}

type CataasAPI interface {
	GetRandomCat() (cat *CatJson, err error)
	BuildUrl(urlPath string, says *string, textSize, width, height *int) *url.URL
}

func (api *cataasAPI) GetRandomCat() (cat *CatJson, err error) {
	uri := *api.baseUrl
	uri.Path = pathCat
	query := url.Values{}
	query.Set("json", "true")
	uri.RawQuery = query.Encode()

	resp, err := api.client.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &cat)
	if err != nil {
		return nil, err
	}
	return
}
func (api *cataasAPI) BuildUrl(urlPath string, says *string, textSize, width, height *int) *url.URL {
	uri := *api.baseUrl
	uri.Path = urlPath
	query := url.Values{}
	if says != nil {
		uri.Path = path.Join(uri.Path, fmt.Sprintf(pathSays, *says))
		if textSize != nil {
			query.Set("size", strconv.Itoa(*textSize))
		}
	}
	if width != nil {
		query.Set("width", strconv.Itoa(*width))
	}
	if height != nil {
		query.Set("height", strconv.Itoa(*height))
	}
	uri.RawQuery = query.Encode()
	return &uri
}

func CreateAPI() CataasAPI {
	baseUrl := &url.URL{
		Host:   "cataas.com",
		Scheme: "https",
	}
	client := http.Client{
		Timeout: defaultTimeout,
	}
	return &cataasAPI{baseUrl, &client}
}
