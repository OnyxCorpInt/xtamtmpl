package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// RecordTypeSecret is the name of the Secret record type
const RecordTypeSecret = "Secret"

// RecordTypeCertificate is the name of the Certificate record type
const RecordTypeCertificate = "Certificate"

// RestAPI is used to make authenticated requests against XTAM. The supplied URL should point to the
type RestAPI struct {
	URL           string
	Authenticator RequestAuthenticator
}

// RequestAuthenticator should execute a REST API call while attaching any information
// needed to authenticate the request.
type RequestAuthenticator interface {
	do(*http.Request) (*http.Response, error)
}

// ListFolder obtains references (name, id, and type) to records in the folder with the given ID.
func (api *RestAPI) ListFolder(folderID string) ([]FolderEntry, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/folder/list/%s", api.URL, folderID), nil)
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

// UnlockSecret returns the value of a Secret record.
func (api *RestAPI) UnlockSecret(secretID int) (string, error) {
	record, err := api.unlockRecord(secretID)
	if err != nil {
		return "", err
	}

	if record.RecordType.Name != RecordTypeSecret {
		return "", fmt.Errorf("wanted Secret but got %s", record.RecordType.Name)
	}

	// So the secret is actually in a JSON document embedded in the first JSON document
	var secret secret
	json.Unmarshal([]byte(record.Custom), &secret)

	return secret.Value, nil
}

// UnlockCertificate returns a PEM-encoded certificate, including, where applicable, the key, and one or more certificates in a chain
func (api *RestAPI) UnlockCertificate(certificateID int) (string, error) {
	record, err := api.unlockRecord(certificateID)
	if err != nil {
		return "", err
	}

	if record.RecordType.Name != RecordTypeCertificate {
		return "", fmt.Errorf("wanted Certificate but got %s", record.RecordType.Name)
	}

	// the certificate is a base64 encoded value in a JSON value, embedded in an outer JSON document
	var cert cert
	json.Unmarshal([]byte(record.Custom), &cert)

	// identifies base64 encoded data with any content type (type could be nothing, or application/x-x509-ca-cert, or possibly others)
	base64regexp := regexp.MustCompile("^data:[^;]*;base64,")

	encodedData := cert.Cert.Data
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

func (api *RestAPI) unlockRecord(id int) (*record, error) {
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

	var record record
	if err = json.Unmarshal(body, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// FolderEntry is a record reference.
type FolderEntry struct {
	Name       string     `json:"name"`
	ID         int        `json:"id"`
	RecordType RecordType `json:"recordType"`
}

// RecordType holds a record type name (see RecordType* constants for some predefined types)
type RecordType struct {
	Name string `json:"name"`
}

type record struct {
	Custom     string     `json:"custom"`
	RecordType RecordType `json:"recordType"`
}

type secret struct {
	Value string `json:"Secret"`
}

type cert struct {
	Cert struct {
		Data string `json:"Data"`
	} `json:"Cert"`
}
