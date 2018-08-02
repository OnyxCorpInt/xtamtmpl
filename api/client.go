package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const RecordTypeSecret = "Secret"
const RecordTypeCertificate = "Certificate"

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
		return nil, fmt.Errorf("unexpected response %d", resp.StatusCode)
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

func (api *RestApi) UnlockSecret(secretId int) (string, error) {
	record, err := api.unlockRecord(secretId)
	if err != nil {
		return "", err
	}

	if record.RecordType.Name != RecordTypeSecret {
		return "", fmt.Errorf("wanted Secret but got %s", record.RecordType.Name)
	}

	// So the secret is actually in a JSON document embedded in the first JSON document. Cue Inception theme...
	var secret Secret
	json.Unmarshal([]byte(record.Custom), &secret)

	return secret.Value, nil
}

// returns a PEM-encoded certificate, including, where applicable, the key, and one or more certificates in a chain
func (api *RestApi) UnlockCertificate(certificateId int) (string, error) {
	record, err := api.unlockRecord(certificateId)
	if err != nil {
		return "", err
	}

	if record.RecordType.Name != RecordTypeCertificate {
		return "", fmt.Errorf("wanted Certificate but got %s", record.RecordType.Name)
	}

	// cue Inception theme: the certificate is a base64 encoded value in a JSON value, embedded in an outer JSON document
	var cert Cert
	json.Unmarshal([]byte(record.Custom), &cert)

	encodedData := cert.Cert.Data
	// identifies base64 encoded data with any content type (type could be nothing, or application/x-x509-ca-cert, or possibly others)
	base64regexp := regexp.MustCompile("^data:[^;]*;base64,")
	matchLoc := base64regexp.FindStringIndex(encodedData)
	if matchLoc == nil {
		return "", fmt.Errorf("expecting base64 encoded certificate data, got: %s", strings.Split(encodedData, ",")[0])
	}

	certData, err := base64.StdEncoding.DecodeString(encodedData[matchLoc[1]:])
	if err != nil {
		return "", err
	}

	return string(certData), nil
}

func (api *RestApi) unlockRecord(id int) (*Record, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/record/unlock/%d", api.URL, id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.Authenticator.do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var record Record
	if err = json.Unmarshal(body, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

type Record struct {
	Custom     string     `json:"custom"`
	RecordType RecordType `json:"recordType"`
}

type Secret struct {
	Value string `json:"Secret"`
}

type Cert struct {
	Cert struct {
		Data string `json:"Data"`
	} `json:"Cert"`
}

type PemFile struct {
	Data     string
	Filename string
}

type FolderEntry struct {
	Name       string     `json:"name"`
	Id         int        `json:"id"`
	RecordType RecordType `json:"recordType"`
}

type RecordType struct {
	Name string `json:"name"`
}
