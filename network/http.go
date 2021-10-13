package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"github.com/sirupsen/logrus"
)

func SimpleHttpPostForm(reqUrl string, values url.Values, v interface{}) error {
	cli := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: time.Second * 10,
		},
	}

	resp, err := cli.PostForm(reqUrl, values)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("http post form fail, status:%d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	v = string(body)

	return nil
}

/*
	简单的 http json 请求
	reqUrls:请求链接
	values:post body传参
	v: 响应参数的指针
*/
func SimpleHttpPostJson(reqUrl string, values []byte, v interface{}) error {
	cli := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: time.Second * 10,
		},
	}

	resp, err := cli.Post(reqUrl, "application/json", bytes.NewReader(values))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("http post json fail, status:%d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}


	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	return nil
}

type Request struct {
	Url     string
	Method  string // get or post
	Headers map[string]string
	Params  map[string][]string    // get 参数
	Data    map[string]interface{} // post 参数
	Retry   int                    // 重试次数
	Timeout time.Duration          // 超时时间
}

func NewDefaultRequest(url string) *Request {
	headers := map[string]string{"Content-Type": "application/json"}
	return &Request{
		Url:     url,
		Headers: headers,
		Retry:   3,
		Method:  http.MethodGet,
		Timeout: 2 * time.Second,
	}
}

func (c *Request) Get(ctx context.Context, headers map[string]string, params map[string][]string, resp interface{}) error {
	c.Method = http.MethodGet
	c.Headers = headers
	c.Params = params
	return c.send(ctx, resp)
}

func (c *Request) Post(ctx context.Context, headers map[string]string, data map[string]interface{}, resp interface{}) error {
	c.Method = http.MethodPost
	c.Headers = headers
	c.Data = data
	return c.send(ctx, resp)
}

func (c *Request) send(ctx context.Context, v interface{}) error {
	cli := &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: time.Second * 10,
		},
	}
	if len(c.Params) > 0 {
		c.Url += "?" + url.Values(c.Params).Encode()
	}
	var reqBody io.Reader
	if len(c.Data) > 0 {
		b, err := json.Marshal(c.Data)
		if err != nil {
			logrus.WithContext(ctx).Errorf("json marshal data failed, err: %v", err.Error())
			return err
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(c.Method, c.Url, reqBody)
	if err != nil {
		logrus.WithContext(ctx).Errorf("new request failed, err: %v", err.Error())
		return err
	}
	if len(c.Headers) > 0 {
		for k, v := range c.Headers {
			req.Header.Set(k, v)
		}
	}
	retry := c.Retry
	if retry <= 0 {
		retry = 3
	}

	var resp *http.Response

	for i := 0; i < retry; i++ {
		resp, err = cli.Do(req)
		if err != nil {
			logrus.WithContext(ctx).Errorf("do request failed, err: %v", err.Error())
			continue
		}
		if resp.StatusCode != 200 {
			err = errors.New(fmt.Sprintf("http send request failed, status:%d", resp.StatusCode))
			logrus.WithContext(ctx).Errorf("do request failed, err: %v", err.Error())
			continue
		}
		break
	}
	if resp != nil {
		defer resp.Body.Close()
		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.WithContext(ctx).Errorf("read resp body failed, err: %v", err.Error())
			return err
		}
		err = json.Unmarshal(respBody, v)
		if err != nil {
			logrus.WithContext(ctx).Errorf("json unmarshal resp body failed, err: %v", err.Error())
			return err
		}
	}
	return err
}

