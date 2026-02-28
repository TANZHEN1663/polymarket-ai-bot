package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	GammaAPIBaseURL = "https://gamma-api.polymarket.com"
	DataAPIBaseURL  = "https://data-api.polymarket.com"
	CLOBAPIBaseURL  = "https://clob.polymarket.com"
)

type Client struct {
	gammaClient  *resty.Client
	dataClient   *resty.Client
	clobClient   *resty.Client
	apiKey       string
	apiSecret    string
	passphrase   string
	httpClient   *http.Client
}

type ClientConfig struct {
	APIKey       string
	APISecret    string
	Passphrase   string
	Timeout      time.Duration
	RetryCount   int
	RetryWait    time.Duration
}

func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	retryCount := config.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	retryWait := config.RetryWait
	if retryWait == 0 {
		retryWait = 1 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	client := &Client{
		apiKey:     config.APIKey,
		apiSecret:  config.APISecret,
		passphrase: config.Passphrase,
		httpClient: httpClient,
	}

	client.gammaClient = resty.New().
		SetBaseURL(GammaAPIBaseURL).
		SetRetryCount(retryCount)

	client.dataClient = resty.New().
		SetBaseURL(DataAPIBaseURL).
		SetRetryCount(retryCount)

	client.clobClient = resty.New().
		SetBaseURL(CLOBAPIBaseURL).
		SetRetryCount(retryCount)

	if config.APIKey != "" && config.APISecret != "" {
		client.clobClient = client.clobClient.
			SetHeader("POLY_API_KEY", config.APIKey).
			SetHeader("POLY_SECRET", config.APISecret)
		if config.Passphrase != "" {
			client.clobClient = client.clobClient.SetHeader("POLY_PASSPHRASE", config.Passphrase)
		}
	}

	return client
}

func (c *Client) executeRequest(ctx context.Context, client *resty.Client, method, endpoint string, params, body, result interface{}) error {
	req := client.R().SetContext(ctx)

	if params != nil {
		req.SetQueryParamsFromValues(url.Values{})
		if paramMap, ok := params.(map[string]string); ok {
			req.SetQueryParams(paramMap)
		}
	}

	if body != nil {
		req.SetBody(body)
	}

	var resp *resty.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = req.Get(endpoint)
	case http.MethodPost:
		resp, err = req.Post(endpoint)
	case http.MethodDelete:
		resp, err = req.Delete(endpoint)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode(), string(resp.Body()))
	}

	if result != nil {
		if err := json.Unmarshal(resp.Body(), result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) get(ctx context.Context, client *resty.Client, endpoint string, params, result interface{}) error {
	return c.executeRequest(ctx, client, http.MethodGet, endpoint, params, nil, result)
}

func (c *Client) post(ctx context.Context, client *resty.Client, endpoint string, body, result interface{}) error {
	return c.executeRequest(ctx, client, http.MethodPost, endpoint, nil, body, result)
}

func (c *Client) delete(ctx context.Context, client *resty.Client, endpoint string, result interface{}) error {
	return c.executeRequest(ctx, client, http.MethodDelete, endpoint, nil, nil, result)
}

func (c *Client) GetGammaClient() *resty.Client {
	return c.gammaClient
}

func (c *Client) GetDataClient() *resty.Client {
	return c.dataClient
}

func (c *Client) GetCLOBClient() *resty.Client {
	return c.clobClient
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

type PaginationParams struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type PaginationResponse struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}
