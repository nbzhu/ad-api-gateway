package pkg

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	netUrl "net/url"
	"time"
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient(timeout time.Duration) *HttpClient {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          4000,
		MaxIdleConnsPerHost:   3500,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &HttpClient{client: client}
}

func (h *HttpClient) Get(ctx context.Context, url string, queryParams map[string]string, headers map[string]string) ([]byte, int, error) {
	u, err := netUrl.Parse(url)
	if err != nil {
		return nil, 0, err
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (h *HttpClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}
