package corp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"
)

const dingTimeOut = time.Second * 1

// Err 微信返回错误
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Client
type Client struct {
	token      string
	mangotoken string
}

// Result 发送消息返回结果
type Result struct {
	Err
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"infvalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// New
func New(token string, mangotoken string) *Client {
	c := new(Client)
	c.token = token
	c.mangotoken = mangotoken
	return c
}

func (c Client) GetToken() string {
	return c.token
}

func (c Client) GetMangoToken() string {
	return c.mangotoken
}

// MangoMessage 消息主体参数
type MangoMessage struct {
	RoomName string `json:"roomname"`
	Text     string `json:"text"`
}

// Send 发送信息
func (c *Client) Send(touser string, msg string) error {
	var token string
	var resultByte []byte
	var err error
	user := strings.Split(touser, "|")

	if user[0] == "tg" {
		token = c.GetToken()
		tgurl, _ := url.Parse("https://api.telegram.org/bot" + token + "/sendMessage")
		params := url.Values{}
		params.Set("chat_id", user[1])
		params.Set("text", msg)
		tgurl.RawQuery = params.Encode()
		urlPath := tgurl.String()
		resultByte, err = jsonGet(urlPath)
	} else {
		var mongomsg *MangoMessage = new(MangoMessage)
		mongomsg.RoomName = user[1]
		mongomsg.Text = msg
		token = c.GetMangoToken()
		mgurl, _ := url.Parse("https://im.imangoim.com:9091/plugins/xhcodrestapi/v1/apiservice/user" + token + "/sendmessage")
		urlPath := mgurl.String()
		resultByte, err = jsonPost(urlPath, mongomsg)
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
	fmt.Println("query:", string(jsonBody))
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))

	if err != nil {
		logger.Info("ding talk new post request err =>", err)
		return nil, err
	}
	req.Header.Set("Authorization", "eXwdrXrvrjsHDs7F")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ding talk post request err =>", err)
		return nil, err
	}

	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	return ioutil.ReadAll(resp.Body)
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

func getClient() *http.Client {
	// 通过http.Client 中的 DialContext 可以设置连接超时和数据接受超时 （也可以使用Dial, 不推荐）
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				conn, err := net.DialTimeout(network, addr, dingTimeOut) // 设置建立链接超时
				if err != nil {
					return nil, err
				}
				_ = conn.SetDeadline(time.Now().Add(dingTimeOut)) // 设置接受数据超时时间
				return conn, nil
			},
			ResponseHeaderTimeout: dingTimeOut, // 设置服务器响应超时时间
		},
	}
}
