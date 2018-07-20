package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RestApi struct {
	URL           string
	Authenticator RequestAuthenticator
}

type RequestAuthenticator interface {
	do(*http.Request) (*http.Response, error)
}

func (api *RestApi) ListFolder(folderId string) ([]FolderEntry, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/folder/list/%s", api.URL, folderId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.Authenticator.do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unexpected response %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []FolderEntry

	if err = json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

type FolderEntry struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}
