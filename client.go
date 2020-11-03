package steam

import (
	"net/http"
	"time"
)

const (
	baseUrl          = "https://steamcommunity.com"
	loginUrl         = "https://steamcommunity.com/login"
	doLoginUrl       = "https://steamcommunity.com/login/dologin"
	rsaUrl           = "https://steamcommunity.com/login/getrsakey"
	defaultUseragent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
)

type Client struct {
	client      *http.Client
	session     *OAuth
	useragent   string
	credentials *Credentials
	apiKey      string
	timeTip     int64
}

type Credentials struct {
	Username       string
	Password       string
	SharedSecret   string
	IdentitySecret string
}

func NewClient(client *http.Client, useragent string, credentials *Credentials) (*Client, error) {
	if useragent == "" {
		useragent = defaultUseragent
	}

	if err := validateCredentials(credentials); err != nil {
		return nil, err
	}

	return &Client{
		client:      client,
		useragent:   useragent,
		credentials: credentials,
	}, nil
}

func (c *Client) getTimeDiff() int64 {
	return time.Now().Add(time.Duration(c.timeTip - time.Now().Unix())).Unix()
}
