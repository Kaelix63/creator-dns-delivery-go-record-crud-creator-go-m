package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

func (e APIError) Error() string { return e.Code + ": " + e.Message }

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    APIError        `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type DNSClient struct {
	baseURL string
	key     string
	http    *http.Client
}

const canonicalCapability = "dns.record.upsert"

func NewDNSClient(baseURL, key string) *DNSClient {
	return &DNSClient{baseURL: baseURL, key: key, http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *DNSClient) do(ctx context.Context, method, path string, query map[string]string, body any, out any) error {
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = encoded
	}
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?"
		values := make([]string, 0, len(query))
		for k, v := range query {
			values = append(values, neturl.QueryEscape(k)+"="+neturl.QueryEscape(v))
		}
		endpoint += strings.Join(values, "&")
	}
	for attempt := 0; attempt < 4; attempt++ {
		var requestBody io.Reader
		if payload != nil {
			requestBody = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		decodeErr := json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("decode response: %w", decodeErr)
		}
		if !env.OK {
			return env.Error
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration(1<<attempt) * time.Second
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 {
				wait = time.Duration(seconds) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request rate limited")
}

func (c *DNSClient) DomainGet(ctx context.Context, domain string, out any) error {
	return c.do(ctx, http.MethodGet, "/v1/dns/domain/get", map[string]string{"domain": domain}, nil, out)
}

func (c *DNSClient) RecordUpsert(ctx context.Context, record map[string]any, out any) error {
	return c.do(ctx, http.MethodPut, "/v1/dns/record/upsert", nil, record, out)
}
