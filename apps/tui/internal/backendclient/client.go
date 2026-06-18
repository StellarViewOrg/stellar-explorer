package backendclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultTimeout = 5 * time.Second

// Client is a small HTTP client for the tui-indexer read API.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// Option customizes client construction.
type Option func(*Client)

// WithHTTPClient injects a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// New creates a read API client against a tui-indexer base URL.
func New(rawBaseURL string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(rawBaseURL) == "" {
		return nil, errors.New("base URL is required")
	}

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q", rawBaseURL)
	}

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Label returns the configured read API base URL for UI source reporting.
func (c *Client) Label() string {
	if c == nil || c.baseURL == nil {
		return ""
	}
	return c.baseURL.String()
}

// Health fetches the backend health status.
func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	var response HealthResponse
	err := c.getJSON(ctx, "/healthz", &response)
	return response, err
}

// LiveFeedSummary fetches the compact live feed summary currently exposed by tui-indexer.
func (c *Client) LiveFeedSummary(ctx context.Context) (LiveFeedSummaryResponse, error) {
	var response LiveFeedSummaryResponse
	err := c.getJSON(ctx, "/v1/feed/live/summary", &response)
	return response, err
}

// Ledger looks up one ledger by sequence.
func (c *Client) Ledger(ctx context.Context, sequence uint32) (LedgerLookupResponse, error) {
	if sequence == 0 {
		return LedgerLookupResponse{}, errors.New("ledger sequence is required")
	}

	var response LedgerLookupResponse
	err := c.getJSON(ctx, fmt.Sprintf("/v1/ledgers/%d", sequence), &response)
	return response, err
}

// Ledgers lists recent ledgers. When before is set, it returns ledgers older than that sequence.
func (c *Client) Ledgers(ctx context.Context, limit int, before uint32) ([]LedgerSummary, error) {
	var response LedgerListResponse
	endpoint := fmt.Sprintf("/v1/ledgers?limit=%d", normalizeLimit(limit))
	if before > 0 {
		endpoint = fmt.Sprintf("%s&before=%d", endpoint, before)
	}
	err := c.getJSON(ctx, endpoint, &response)
	return response.Ledgers, err
}

// LedgerTransactions lists transactions for one ledger.
func (c *Client) LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]TransactionSummary, error) {
	if sequence == 0 {
		return nil, errors.New("ledger sequence is required")
	}

	var response struct {
		Transactions []TransactionSummary `json:"transactions"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/ledgers/%d/transactions?limit=%d&offset=%d", sequence, normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Transactions, err
}

// Search queries the TUI indexer search endpoint.
func (c *Client) Search(ctx context.Context, query string, limit int) (SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchResponse{}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	var response SearchResponse
	endpoint := fmt.Sprintf("/v1/search?q=%s&limit=%d", url.QueryEscape(query), limit)
	err := c.getJSON(ctx, endpoint, &response)
	return response, err
}

// Transaction looks up one transaction by hash.
func (c *Client) Transaction(ctx context.Context, hash string) (TransactionLookupResponse, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return TransactionLookupResponse{}, errors.New("transaction hash is required")
	}

	var response TransactionLookupResponse
	err := c.getJSON(ctx, "/v1/transactions/"+url.PathEscape(hash), &response)
	return response, err
}

// Account looks up one account by address.
func (c *Client) Account(ctx context.Context, id string) (AccountLookupResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return AccountLookupResponse{}, errors.New("account id is required")
	}

	var response AccountLookupResponse
	err := c.getJSON(ctx, "/v1/accounts/"+url.PathEscape(id), &response)
	return response, err
}

// Accounts lists recently updated accounts.
func (c *Client) Accounts(ctx context.Context, limit int) ([]AccountDetail, error) {
	var response AccountListResponse
	err := c.getJSON(ctx, fmt.Sprintf("/v1/accounts?limit=%d", normalizeLimit(limit)), &response)
	return response.Accounts, err
}

// AccountTransactions lists transactions related to one account.
func (c *Client) AccountTransactions(ctx context.Context, id string, limit int) ([]TransactionSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("account id is required")
	}

	var response struct {
		Transactions []TransactionSummary `json:"transactions"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/accounts/%s/transactions?limit=%d", url.PathEscape(id), normalizeLimit(limit)), &response)
	return response.Transactions, err
}

// AccountOperations lists recent operations related to one account.
func (c *Client) AccountOperations(ctx context.Context, id string, limit int, offset int) ([]OperationSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("account id is required")
	}

	var response struct {
		Operations []OperationSummary `json:"operations"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/accounts/%s/operations?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Operations, err
}

// AccountTimeline lists normalized account activity rows.
func (c *Client) AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]TimelineItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("account id is required")
	}

	var response TimelineResponse
	endpoint := fmt.Sprintf("/v1/accounts/%s/timeline?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset))
	if category = strings.TrimSpace(category); category != "" {
		endpoint += "&type=" + url.QueryEscape(category)
	}
	err := c.getJSON(ctx, endpoint, &response)
	return response.Items, err
}

// Asset looks up one asset by CODE:ISSUER.
func (c *Client) Asset(ctx context.Context, code string, issuer string) (AssetLookupResponse, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return AssetLookupResponse{}, errors.New("asset code and issuer are required")
	}

	var response AssetLookupResponse
	err := c.getJSON(ctx, "/v1/assets/"+url.PathEscape(code+":"+issuer), &response)
	return response, err
}

// Assets lists recently updated assets.
func (c *Client) Assets(ctx context.Context, limit int) ([]AssetDetail, error) {
	var response AssetListResponse
	err := c.getJSON(ctx, fmt.Sprintf("/v1/assets?limit=%d", normalizeLimit(limit)), &response)
	return response.Assets, err
}

// AssetTransactions lists transactions related to one asset.
func (c *Client) AssetTransactions(ctx context.Context, code string, issuer string, limit int) ([]TransactionSummary, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return nil, errors.New("asset code and issuer are required")
	}

	var response struct {
		Transactions []TransactionSummary `json:"transactions"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/assets/%s/transactions?limit=%d", url.PathEscape(code+":"+issuer), normalizeLimit(limit)), &response)
	return response.Transactions, err
}

// AssetHolders lists top holders for one asset.
func (c *Client) AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]AssetHolderSummary, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return nil, errors.New("asset code and issuer are required")
	}

	var response struct {
		Holders []AssetHolderSummary `json:"holders"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/assets/%s/holders?limit=%d&offset=%d", url.PathEscape(code+":"+issuer), normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Holders, err
}

// AssetTimeline lists normalized asset activity rows.
func (c *Client) AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]TimelineItem, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return nil, errors.New("asset code and issuer are required")
	}

	var response TimelineResponse
	endpoint := fmt.Sprintf("/v1/assets/%s/timeline?limit=%d&offset=%d", url.PathEscape(code+":"+issuer), normalizeLimit(limit), normalizeOffset(offset))
	if category = strings.TrimSpace(category); category != "" {
		endpoint += "&type=" + url.QueryEscape(category)
	}
	err := c.getJSON(ctx, endpoint, &response)
	return response.Items, err
}

// Contract looks up one contract by contract ID.
func (c *Client) Contract(ctx context.Context, id string) (ContractLookupResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ContractLookupResponse{}, errors.New("contract id is required")
	}

	var response ContractLookupResponse
	err := c.getJSON(ctx, "/v1/contracts/"+url.PathEscape(id), &response)
	return response, err
}

// ContractSpec fetches the structured Soroban contract spec for one contract.
func (c *Client) ContractSpec(ctx context.Context, id string) (ContractSpec, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ContractSpec{}, errors.New("contract id is required")
	}

	var response ContractSpec
	err := c.getJSON(ctx, "/v1/contracts/"+url.PathEscape(id)+"/spec", &response)
	return response, err
}

// Contracts lists recently updated contracts.
func (c *Client) Contracts(ctx context.Context, limit int) ([]ContractDetail, error) {
	var response ContractListResponse
	err := c.getJSON(ctx, fmt.Sprintf("/v1/contracts?limit=%d", normalizeLimit(limit)), &response)
	return response.Contracts, err
}

// ContractTransactions lists recent transactions for one contract.
func (c *Client) ContractTransactions(ctx context.Context, id string, limit int) ([]TransactionSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("contract id is required")
	}

	var response struct {
		Transactions []TransactionSummary `json:"transactions"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/contracts/%s/transactions?limit=%d", url.PathEscape(id), normalizeLimit(limit)), &response)
	return response.Transactions, err
}

// ContractEvents lists recent contract events for one contract.
func (c *Client) ContractEvents(ctx context.Context, id string, limit int, offset int) ([]ContractEventSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("contract id is required")
	}

	var response struct {
		Events []ContractEventSummary `json:"events"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/contracts/%s/events?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Events, err
}

// ContractInvocations lists recent invoke_host_function operations for one contract.
func (c *Client) ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]OperationSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("contract id is required")
	}

	var response struct {
		Operations []OperationSummary `json:"operations"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/contracts/%s/invocations?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Operations, err
}

// ContractStorage lists storage entries for one contract.
func (c *Client) ContractStorage(ctx context.Context, id string, limit int, offset int) ([]ContractStorageSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("contract id is required")
	}

	var response struct {
		Storage []ContractStorageSummary `json:"storage"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/v1/contracts/%s/storage?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset)), &response)
	return response.Storage, err
}

// ContractTimeline lists normalized contract activity rows.
func (c *Client) ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]TimelineItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("contract id is required")
	}

	var response TimelineResponse
	endpoint := fmt.Sprintf("/v1/contracts/%s/timeline?limit=%d&offset=%d", url.PathEscape(id), normalizeLimit(limit), normalizeOffset(offset))
	if category = strings.TrimSpace(category); category != "" {
		endpoint += "&type=" + url.QueryEscape(category)
	}
	err := c.getJSON(ctx, endpoint, &response)
	return response.Items, err
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func (c *Client) getJSON(ctx context.Context, endpoint string, dest interface{}) error {
	requestURL := *c.baseURL
	endpointPath, rawQuery, _ := strings.Cut(endpoint, "?")
	requestURL.Path = path.Join(c.baseURL.Path, endpointPath)
	requestURL.RawQuery = rawQuery

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return decodeHTTPError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func decodeHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))

	var apiErr APIError
	if len(body) > 0 && json.Unmarshal(body, &apiErr) == nil && strings.TrimSpace(apiErr.Error) != "" {
		return HTTPError{
			StatusCode: resp.StatusCode,
			Message:    apiErr.Error,
		}
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	return HTTPError{
		StatusCode: resp.StatusCode,
		Message:    message,
	}
}
