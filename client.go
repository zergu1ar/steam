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

	LanguageEng = "english"
	LanguageRus = "russian"
)

type Client struct {
	client      *http.Client
	session     *OAuth
	useragent   string
	credentials *Credentials
	apiKey      string
	timeTip     int64
	language    string
}

type Credentials struct {
	Username       string
	Password       string
	SharedSecret   string
	IdentitySecret string
}

func NewClient(client *http.Client, useragent string, language string, credentials *Credentials) (*Client, error) {
	if useragent == "" {
		useragent = defaultUseragent
	}
	if language == "" {
		language = LanguageEng
	}

	if err := validateCredentials(credentials); err != nil {
		return nil, err
	}

	return &Client{
		client:      client,
		useragent:   useragent,
		credentials: credentials,
		language:    language,
	}, nil
}

func (c *Client) getTimeDiff() int64 {
	return time.Now().Add(time.Duration(c.timeTip - time.Now().Unix())).Unix()
}

func (c *Client) GetSteamId() SteamID {
	if c.session != nil {
		return c.session.SteamID
	}
	return SteamID(0)
}
