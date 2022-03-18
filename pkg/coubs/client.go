package coubs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	cookies *Cookies
}

func NewClient(cookies *Cookies) *Client {
	return &Client{
		cookies: cookies,
	}
}

func (c *Client) ChannelTimeline(name string, page int) (*PageResponse, error) {
	st, err := c.cookies.Get()
	if err != nil {
		return nil, err
	}
	req := st.Request

	requestURL, err := url.Parse(
		fmt.Sprintf("https://coub.com/api/v2/timeline/channel/%s?order_by=newest&permalink=%s&type=&page=%d",
			name, name, page,
		))
	if err != nil {
		return nil, err
	}

	req.Method = "GET"
	req.URL = requestURL
	req.RequestURI = ""
	req.Header.Del("Accept-Encoding")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = c.cookies.Update(req, resp)
	if err != nil {
		return nil, err
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pageResponse PageResponse
	err = json.Unmarshal(body, &pageResponse)
	if err != nil {
		return nil, err
	}

	return &pageResponse, nil
}
