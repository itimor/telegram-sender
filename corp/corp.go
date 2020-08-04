package corp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Err 微信返回错误
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Client
type Client struct {
	token   string
	openUrl string
}

// Result 发送消息返回结果
type Result struct {
	Err
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"infvalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// New
func New(token string) *Client {
	c := new(Client)
	c.openUrl = "https://api.telegram.org/bot"
	c.token = token
	return c
}

func (c Client) GetToken() string {
	return c.token
}

// Send 发送信息
func (c *Client) Send(touser string, msg string) error {
	token := c.GetToken()

	var method string = "GET"
	var resultByte []byte
	var err error
	if method == "GET" {
		URL, _ := url.Parse(c.openUrl + token + "/sendMessage")
		params := url.Values{}
		params.Set("chat_id", touser)
		params.Set("text", msg)
		URL.RawQuery = params.Encode()
		urlPath := URL.String()
		println(urlPath)
		resultByte, err = jsonGet(urlPath)
	} else {
		// 此处无用代码
		url := c.openUrl + token + "/sendMessage?chat_id=" + touser + "&text=" + msg
		resultByte, err = jsonPost(url, msg)
	}

	if err != nil {
		return fmt.Errorf("invoke send api fail: %v", err)
	}

	result := Result{}
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return fmt.Errorf("parse send api response fail: %v", err)
	}

	if result.ErrCode != 0 {
		err = fmt.Errorf("invoke send api return ErrCode = %d", result.ErrCode)
	}

	if result.InvalidUser != "" || result.InvalidParty != "" || result.InvalidTag != "" {
		err = fmt.Errorf("invoke send api partial fail, invalid user: %s, invalid party: %s, invalid tag: %s", result.InvalidUser, result.InvalidParty, result.InvalidTag)
	}

	return err
}

// transport 全局复用，提升性能
var transport = &http.Transport{
	TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	DisableCompression: true,
}

func jsonGet(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if r.Body == nil {
		return nil, fmt.Errorf("response body of %s is nil", url)
	}

	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

func jsonPost(url string, data interface{}) ([]byte, error) {
	jsonBody, err := encodeJSON(data)
	if err != nil {
		return nil, err
	}

	r, err := http.Post(url, "application/json;charset=utf-8", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	if r.Body == nil {
		return nil, fmt.Errorf("response body of %s is nil", url)
	}

	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

func encodeJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
