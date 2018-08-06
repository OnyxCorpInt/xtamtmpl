package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
)

// CASAuth obtains a service ticket from CASURL for BaseURL using User and Password as credentials.
type CASAuth struct {
	CASURL   string
	BaseURL  string
	User     string
	Password string

	client http.Client
}

func (cas *CASAuth) do(req *http.Request) (*http.Response, error) {
	baseURL, err := url.Parse(cas.BaseURL)
	if err != nil {
		return nil, err
	}

	// unless we specify an in-memory cookie store, the client won't retain cookies set in responses
	if cas.client.Jar == nil {
		cas.client.Jar, err = cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
	}

	// with CAS authentication, we use session cookies. So the absence of cookies means we must authenticate.
	if len(cas.client.Jar.Cookies(baseURL)) == 0 {
		if err = cas.authenticate(); err != nil {
			return nil, err
		}

	}

	// now that we've authenticated cas.client, we can complete the original request
	return cas.client.Do(req)
}

func (cas *CASAuth) authenticate() error {
	tgtResp, err := cas.client.PostForm(cas.CASURL+"/v1/tickets", url.Values{"username": []string{cas.User}, "password": []string{cas.Password}})
	if err != nil {
		return err
	}

	if tgtResp.StatusCode != http.StatusCreated {
		rb, _ := httputil.DumpResponse(tgtResp, true)
		fmt.Println(string(rb))
		return fmt.Errorf("unexpected status code while obaining TGT: %d", tgtResp.StatusCode)
	}

	tgtLocation := tgtResp.Header.Get("Location")
	if tgtLocation == "" {
		return errors.New("authentication failure (unable to obtain TGT location from CAS)")
	}

	stResp, err := cas.client.PostForm(tgtLocation, url.Values{"service": []string{cas.BaseURL + "/"}})
	if err != nil {
		return err
	}

	if stResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while service ticket: %d", stResp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(stResp.Body)
	if err != nil {
		return err
	}

	serviceTicket := string(bodyBytes)

	if serviceTicket == "" {
		return errors.New("unable to obtain service ticket from " + tgtLocation)
	}

	// calling for side-effect of populating auth.client.Jar
	cookieResp, err := cas.client.Get(cas.BaseURL + "?ticket=" + url.QueryEscape(serviceTicket))
	if cookieResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from cookie request: %d", cookieResp.StatusCode)
	}
	return err
}
