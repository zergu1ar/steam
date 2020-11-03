package steam

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (confirmation *Confirmation) Answer(client *Client, answer string) error {
	return client.AnswerConfirmation(confirmation, answer)
}

type ConfirmationAnswerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *Client) confirmationReqWorker(delay int64) {
	for {
		select {
		case req := <-c.requestQueue["confirmation"]:
			req.ResponseChan <- c.execConfirmationRequest(req.Url, req.Body, req.Params, nil)
			time.Sleep(time.Second * time.Duration(delay))
		case <-c.ctx.Done():
			close(c.requestQueue["confirmation"])
			return
		}
	}
}

func (c *Client) GetConfirmations() ([]*Confirmation, error) {
	req := RequestItem{
		Url: "conf?",
		Params: url.Values{
			"tag": {"confirmation"},
		},
		ResponseChan: make(chan RequestResponse),
	}
	c.requestQueue["confirmation"] <- req

	resp := <-req.ResponseChan
	close(req.ResponseChan)
	if resp.Error != nil {
		return nil, resp.Error
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
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

func (c *Client) execConfirmationRequest(uri string, body io.Reader, params url.Values, values map[string]interface{}) RequestResponse {
	key, err := GenerateConfirmationCode(c.credentials.IdentitySecret, "confirmation", c.getTimeDiff())
	if err != nil {
		return RequestResponse{
			Error:  err,
			Body:   []byte(""),
			Status: http.StatusBadRequest,
		}
	}

	params.Set("p", c.session.DeviceID)
	params.Set("a", c.session.SteamID.ToString())
	params.Set("t", strconv.FormatInt(c.getTimeDiff(), 10))
	params.Set("m", "android")
	params.Set("k", key)

	for k, v := range values {
		switch v := v.(type) {
		case string:
			params.Add(k, v)
		case uint64:
			params.Add(k, strconv.FormatUint(v, 10))
		}
	}

	respBody := []byte("")
	resp, err := c.client.Get(confirmationUrl + uri + params.Encode())
	if err != nil {
		return RequestResponse{
			Error:  err,
			Body:   respBody,
			Status: http.StatusBadRequest,
		}
	}

	respBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return RequestResponse{
		Error:  err,
		Body:   respBody,
		Status: resp.StatusCode,
	}
}

func (c *Client) AnswerConfirmation(confirmation *Confirmation, answer string) error {
	op := map[string]interface{}{
		"op":  answer,
		"cid": confirmation.ID,
		"ck":  confirmation.Key,
	}

	req := RequestItem{
		Url:  "ajaxop?",
		Body: nil,
		Params: url.Values{
			"tag": {answer},
		},
		ResponseChan: make(chan RequestResponse),
		Values:       op,
	}

	c.requestQueue["confirmation"] <- req

	resp := <-req.ResponseChan

	if resp.Error != nil {
		return resp.Error
	}

	var response ConfirmationAnswerResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}
