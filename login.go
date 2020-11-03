package steam

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type LoginResponse struct {
	Success      bool   `json:"success"`
	PublicKeyMod string `json:"publickey_mod"`
	PublicKeyExp string `json:"publickey_exp"`
	Timestamp    string `json:"timestamp"`
	TokenGID     string `json:"token_gid"`
}

type LoginSession struct {
	Success           bool   `json:"success"`
	LoginComplete     bool   `json:"login_complete"`
	RequiresTwoFactor bool   `json:"requires_twofactor"`
	Message           string `json:"message"`
	RedirectURI       string `json:"redirect_uri"`
	OAuth             OAuth  `json:"transfer_parameters"`
}

type OAuth struct {
	ID          string  `json:"-"`
	DeviceID    string  `json:"-"`
	SteamID     SteamID `json:"steamid,string"`
	Auth        string  `json:"auth"`
	TokenSecure string  `json:"token_secure"`
	WebCookie   string  `json:"webcookie"`
}

func (c *Client) Login() error {
	err := c.setupCookie()
	if err != nil {
		return err
	}

	response, err := c.makeLoginRequest(c.credentials.Username)
	if err != nil {
		return err
	}

	var twoFactorCode string
	if len(c.credentials.SharedSecret) != 0 {
		if twoFactorCode, err = GenerateTwoFactorCode(c.credentials.SharedSecret, c.getTimeDiff()); err != nil {
			return err
		}
	}

	return c.proceedDirectLogin(response, c.credentials.Username, c.credentials.Password, twoFactorCode)
}

func (c *Client) setupCookie() error {
	req, err := http.NewRequest(
		http.MethodGet,
		loginUrl,
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	steamUrl, err := url.Parse(baseUrl)
	if err != nil {
		return err
	}

	now := time.Now()
	_, offset := now.Zone()

	cookies := []*http.Cookie{
		{Name: "timezoneOffset", Value: fmt.Sprintf("%d,0", offset)},
	}

	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, &http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}

	jar.SetCookies(steamUrl, cookies)
	c.client.Jar = jar

	return nil
}

func (c *Client) makeLoginRequest(accountName string) (*LoginResponse, error) {
	reqData := url.Values{
		"username":   {accountName},
		"donotcache": {strconv.FormatInt(time.Now().Unix()*1000, 10)},
	}.Encode()

	req, err := http.NewRequest(
		http.MethodPost,
		rsaUrl,
		strings.NewReader(reqData),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(reqData)))
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Origin", baseUrl)
	req.Header.Add("Referer", loginUrl)
	req.Header.Add("User-Agent", c.useragent)
	req.Header.Add("Accept", "*/*")

	resp, err := c.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	var response LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, InvalidCredentialsError
	}

	return &response, nil
}

func (c *Client) proceedDirectLogin(response *LoginResponse, accountName, password, twoFactorCode string) error {
	var n big.Int
	n.SetString(response.PublicKeyMod, 16)

	exp, err := strconv.ParseInt(response.PublicKeyExp, 16, 32)
	if err != nil {
		return err
	}

	pub := rsa.PublicKey{N: &n, E: int(exp)}
	rsaOut, err := rsa.EncryptPKCS1v15(rand.Reader, &pub, []byte(password))
	if err != nil {
		return err
	}

	reqData := url.Values{
		"captcha_text":      {""},
		"captchagid":        {"-1"},
		"emailauth":         {""},
		"emailsteamid":      {""},
		"username":          {accountName},
		"password":          {base64.StdEncoding.EncodeToString(rsaOut)},
		"remember_login":    {"true"},
		"rsatimestamp":      {response.Timestamp},
		"twofactorcode":     {twoFactorCode},
		"donotcache":        {strconv.FormatInt(time.Now().Unix()*1000, 10)},
		"loginfriendlyname": {""},
	}.Encode()

	req, err := http.NewRequest(
		http.MethodPost,
		doLoginUrl,
		strings.NewReader(reqData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(reqData)))
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Origin", baseUrl)
	req.Header.Add("Referer", loginUrl)
	req.Header.Add("User-Agent", c.useragent)
	req.Header.Add("Accept", "*/*")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	loginSession := &LoginSession{}
	if err := json.NewDecoder(resp.Body).Decode(&loginSession); err != nil {
		return err
	}

	if !loginSession.Success {
		if loginSession.RequiresTwoFactor {
			return RequireTwoFactorError
		}

		return errors.New(loginSession.Message)
	}

	steamUrl, _ := url.Parse(baseUrl)
	cookies := c.client.Jar.Cookies(steamUrl)
	for _, cookie := range cookies {
		if cookie.Name == "sessionid" {
			loginSession.OAuth.ID = cookie.Value
			break
		}
	}

	c.session = &loginSession.OAuth

	if c.session.ID == "" {
		return InvalidSessionError
	}

	sum := md5.Sum([]byte(accountName + password))
	c.session.DeviceID = fmt.Sprintf(
		"android:%x-%x-%x-%x-%x",
		sum[:2], sum[2:4], sum[4:6], sum[6:8], sum[8:10],
	)

	return nil
}
