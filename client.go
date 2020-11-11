package steam

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
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

	confirmationDelay = 3
)

type Client struct {
	ctx          context.Context
	client       *http.Client
	session      *OAuth
	useragent    string
	credentials  *Credentials
	apiKey       string
	timeDiff     int64
	language     string
	Destroy      func()
	requestQueue map[string]chan RequestItem
}

type Credentials struct {
	Username       string
	Password       string
	SharedSecret   string
	IdentitySecret string
}

type (
	RequestItem struct {
		Url          string
		Body         io.Reader
		Params       url.Values
		ResponseChan chan RequestResponse
		Values       map[string]interface{}
	}
	RequestResponse struct {
		Error  error
		Body   []byte
		Status int
	}
)

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

	ctx, cancel := context.WithCancel(context.Background())

	queue := make(map[string]chan RequestItem)
	queue["confirmation"] = make(chan RequestItem, 1000)

	steamClient := &Client{
		ctx:          ctx,
		Destroy:      cancel,
		client:       client,
		useragent:    useragent,
		credentials:  credentials,
		language:     language,
		requestQueue: queue,
	}

	timeTip, err := GetTimeTip()
	if err != nil {
		log.Fatal(err)
	}
	steamClient.timeDiff = timeTip.Time - time.Now().Unix()

	// start goroutines to perform requests
	go steamClient.confirmationReqWorker(confirmationDelay)
	go steamClient.checkSession()

	return steamClient, nil
}

func (c *Client) getTimeDiff() int64 {
	return time.Now().Unix() + c.timeDiff
}

func (c *Client) GetSteamId() SteamID {
	if c.session != nil {
		return c.session.SteamID
	}
	return SteamID(0)
}
