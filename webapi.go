package steam

import (
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	apiKeyURL           = "https://steamcommunity.com/dev/apikey"
	accessDeniedPattern = "<h2>Access Denied</h2>"
)

var (
	keyRegExp = regexp.MustCompile("<p>Key: ([0-9A-F]+)</p>")
)

func (c *Client) GetWebAPIKey() (string, error) {
	resp, err := c.client.Get(apiKeyURL)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return "", err
	}

	return c.parseKey(resp)
}

func (c *Client) parseKey(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if m, err := regexp.Match(accessDeniedPattern, body); err != nil {
		return "", err
	} else if m {
		return "", ApiAccessDeniedError
	}

	submatch := keyRegExp.FindStringSubmatch(string(body))
	if len(submatch) != 2 {
		return "", ApiKeyNotFoundError
	}

	c.apiKey = submatch[1]
	return submatch[1], nil
}
