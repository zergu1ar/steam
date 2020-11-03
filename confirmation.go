package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	confirmationUrl = "https://steamcommunity.com/mobileconf/"
	AnswerAllow     = "allow"
	AnswerDeny      = "deny"
)

type Confirmation struct {
	ID        uint64
	Key       uint64
	Title     string
	Receiving string
	Since     string
	OfferID   uint64
}

type ConfirmationAnswerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (confirmation *Confirmation) Answer(client *Client, answer string) error {
	return client.AnswerConfirmation(confirmation, answer)
}

func (c *Client) GetConfirmations() ([]*Confirmation, error) {
	key, err := GenerateConfirmationCode(c.credentials.IdentitySecret, "confirmation", c.getTimeDiff())
	if err != nil {
		return nil, err
	}

	resp, err := c.execConfirmationRequest("conf?", key, "confirmation", c.getTimeDiff(), nil)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(io.Reader(resp.Body))
	if err != nil {
		return nil, err
	}

	entries := doc.Find(".mobileconf_list_entry")
	if entries == nil {
		return nil, ConfirmationsNotFoundError
	}

	descriptions := doc.Find(".mobileconf_list_entry_description")
	if descriptions == nil {
		return nil, ConfirmationsDescriptionNotFoundError
	}

	confirmations := make([]*Confirmation, 0)
	for k, sel := range entries.Nodes {
		confirmation := &Confirmation{}
		for _, attr := range sel.Attr {
			if attr.Key == "data-confid" {
				confirmation.ID, _ = strconv.ParseUint(attr.Val, 10, 64)
			} else if attr.Key == "data-key" {
				confirmation.Key, _ = strconv.ParseUint(attr.Val, 10, 64)
			} else if attr.Key == "data-creator" {
				confirmation.OfferID, _ = strconv.ParseUint(attr.Val, 10, 64)
			}
		}

		descSel := descriptions.Nodes[k]
		depth := 0
		for child := descSel.FirstChild; child != nil; child = child.NextSibling {
			for n := child.FirstChild; n != nil; n = n.NextSibling {
				switch depth {
				case 0:
					confirmation.Title = n.Data
				case 1:
					confirmation.Receiving = n.Data
				case 2:
					confirmation.Since = n.Data
				}
				depth++
			}
		}

		confirmations = append(confirmations, confirmation)
	}

	return confirmations, nil
}

func (c *Client) execConfirmationRequest(request, key, tag string, current int64, values map[string]interface{}) (*http.Response, error) {
	params := url.Values{
		"p":   {c.session.DeviceID},
		"a":   {c.session.SteamID.ToString()},
		"k":   {key},
		"t":   {strconv.FormatInt(current, 10)},
		"m":   {"android"},
		"tag": {tag},
	}

	for k, v := range values {
		switch v := v.(type) {
		case string:
			params.Add(k, v)
		case uint64:
			params.Add(k, strconv.FormatUint(v, 10))
		default:
			return nil, fmt.Errorf("execConfirmationRequest: missing implementation for type %v", v)
		}
	}
	return c.client.Get(confirmationUrl + request + params.Encode())
}

func (c *Client) AnswerConfirmation(confirmation *Confirmation, answer string) error {
	key, err := GenerateConfirmationCode(c.credentials.IdentitySecret, answer, c.getTimeDiff())
	if err != nil {
		return err
	}

	op := map[string]interface{}{
		"op":  answer,
		"cid": confirmation.ID,
		"ck":  confirmation.Key,
	}

	resp, err := c.execConfirmationRequest("ajaxop?", key, answer, c.getTimeDiff(), op)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}

	var response ConfirmationAnswerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}
