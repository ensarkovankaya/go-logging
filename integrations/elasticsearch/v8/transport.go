package elastic

import "net/http"

var DefaultTransport *Transport

type Transport struct {
	Client *http.Client
}

func (t *Transport) Perform(req *http.Request) (*http.Response, error) {
	return t.Client.Do(req)
}

func init() {
	DefaultTransport = &Transport{
		Client: &http.Client{},
	}
}
