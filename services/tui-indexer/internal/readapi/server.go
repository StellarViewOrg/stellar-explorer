package readapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

type Server struct {
	store readStore
	rpc   *source.RPCClient
	log   *log.Logger
}

type readStore interface {
	Ping(ctx context.Context) error
	GetLastIngestedLedger(ctx context.Context) (uint32, error)
	GetLatestLedgerSummary(ctx context.Context) (*store.ReadLedgerSummary, error)
	ListLedgerSummaries(ctx context.Context, limit int, before uint32) ([]store.ReadLedgerSummary, error)
	ListRecentTransactionSummaries(ctx context.Context, limit int) ([]store.ReadTransactionSummary, error)
	Search(ctx context.Context, query string, limit int) ([]store.ReadSearchResult, error)
	GetLedgerSummaryBySequence(ctx context.Context, sequence uint32) (*store.ReadLedgerSummary, error)
	ListTransactionSummariesByLedger(ctx context.Context, sequence uint32, limit int, offset int) ([]store.ReadTransactionSummary, error)
	GetTransactionByHash(ctx context.Context, hash string) (*store.ReadTransactionDetail, error)
	ListOperationsByTransactionHash(ctx context.Context, hash string) ([]store.ReadOperationSummary, error)
	ListEffectsByTransactionHash(ctx context.Context, hash string) ([]store.ReadEffectSummary, error)
	GetAccountByID(ctx context.Context, id string) (*store.ReadAccountDetail, error)
	ListAccountDetails(ctx context.Context, limit int) ([]store.ReadAccountDetail, error)
	ListTrustlinesByAccountID(ctx context.Context, accountID string) ([]store.ReadTrustlineSummary, error)
	ListAccountSignersByAccountID(ctx context.Context, accountID string) ([]store.ReadAccountSignerSummary, error)
	ListTransactionSummariesByAccount(ctx context.Context, accountID string, limit int) ([]store.ReadTransactionSummary, error)
	ListOperationSummariesByAccount(ctx context.Context, accountID string, limit int, offset int) ([]store.ReadOperationSummary, error)
	ListAccountTimeline(ctx context.Context, accountID string, limit int, offset int, category string) ([]store.ReadTimelineItem, error)
	GetAssetByCodeIssuer(ctx context.Context, code string, issuer string) (*store.ReadAssetDetail, error)
	ListAssetDetails(ctx context.Context, limit int) ([]store.ReadAssetDetail, error)
	ListAssetHoldersByCodeIssuer(ctx context.Context, code string, issuer string, limit int, offset int) ([]store.ReadAssetHolderSummary, error)
	ListTransactionSummariesByAsset(ctx context.Context, code string, issuer string, limit int) ([]store.ReadTransactionSummary, error)
	ListAssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]store.ReadTimelineItem, error)
	GetContractByID(ctx context.Context, contractID string) (*store.ReadContractDetail, error)
	GetContractSpecByID(ctx context.Context, contractID string) (*store.ReadContractSpec, error)
	ListContractDetails(ctx context.Context, limit int) ([]store.ReadContractDetail, error)
	ListTransactionSummariesByContract(ctx context.Context, contractID string, limit int) ([]store.ReadTransactionSummary, error)
	ListContractEventSummariesByContractID(ctx context.Context, contractID string, limit int, offset int) ([]store.ReadContractEventSummary, error)
	ListContractStorageByContractID(ctx context.Context, contractID string, limit int, offset int) ([]store.ReadContractStorageSummary, error)
	ListOperationSummariesByContract(ctx context.Context, contractID string, limit int, offset int) ([]store.ReadOperationSummary, error)
	ListContractTimeline(ctx context.Context, contractID string, limit int, offset int, category string) ([]store.ReadTimelineItem, error)
}

func NewServer(db readStore, rpc *source.RPCClient, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}

	return &Server{
		store: db,
		rpc:   rpc,
		log:   logger,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/feed/live/summary", s.handleLiveFeedSummary)
	mux.HandleFunc("/v1/search", s.handleSearch)
	mux.HandleFunc("/v1/accounts", s.handleAccountList)
	mux.HandleFunc("/v1/accounts/", s.handleAccountLookup)
	mux.HandleFunc("/v1/assets", s.handleAssetList)
	mux.HandleFunc("/v1/assets/", s.handleAssetLookup)
	mux.HandleFunc("/v1/contracts", s.handleContractList)
	mux.HandleFunc("/v1/contracts/", s.handleContractLookup)
	mux.HandleFunc("/v1/ledgers", s.handleLedgerList)
	mux.HandleFunc("/v1/ledgers/", s.handleLedgerLookup)
	mux.HandleFunc("/v1/transactions/", s.handleTransactionLookup)
	return mux
}

type healthResponse struct {
	Status             string  `json:"status"`
	Database           string  `json:"database"`
	RPC                string  `json:"rpc"`
	LatestRPCLedger    *uint32 `json:"latest_rpc_ledger,omitempty"`
	LastIngestedLedger uint32  `json:"last_ingested_ledger"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := healthResponse{
		Status:   "ok",
		Database: "ok",
		RPC:      "disabled",
	}

	if err := s.store.Ping(ctx); err != nil {
		resp.Status = "error"
		resp.Database = err.Error()
		writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}

	lastIngested, err := s.store.GetLastIngestedLedger(ctx)
	if err != nil {
		resp.Status = "error"
		resp.Database = err.Error()
		writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}
	resp.LastIngestedLedger = lastIngested

	if s.rpc != nil {
		latest, err := s.rpc.GetLatestLedger(ctx)
		if err != nil {
			resp.Status = "degraded"
			resp.RPC = err.Error()
			writeJSON(w, http.StatusServiceUnavailable, resp)
			return
		}
		resp.RPC = "ok"
		resp.LatestRPCLedger = &latest.Sequence
	}

	writeJSON(w, http.StatusOK, resp)
}

type liveFeedSummaryResponse struct {
	LastIngestedLedger uint32                         `json:"last_ingested_ledger"`
	LatestLedger       *store.ReadLedgerSummary       `json:"latest_ledger,omitempty"`
	RecentTransactions []store.ReadTransactionSummary `json:"recent_transactions"`
}

func (s *Server) handleLiveFeedSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	lastIngested, err := s.store.GetLastIngestedLedger(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	latestLedger, err := s.store.GetLatestLedgerSummary(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentTransactions, err := s.store.ListRecentTransactionSummaries(ctx, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, liveFeedSummaryResponse{
		LastIngestedLedger: lastIngested,
		LatestLedger:       latestLedger,
		RecentTransactions: recentTransactions,
	})
}

type transactionLookupResponse struct {
	Transaction *store.ReadTransactionDetail `json:"transaction"`
	Operations  []store.ReadOperationSummary `json:"operations"`
	Effects     []store.ReadEffectSummary    `json:"effects"`
}

type ledgerLookupResponse struct {
	Ledger       *store.ReadLedgerSummary       `json:"ledger"`
	Transactions []store.ReadTransactionSummary `json:"transactions"`
}

type transactionListResponse struct {
	Transactions []store.ReadTransactionSummary `json:"transactions"`
}

type ledgerListResponse struct {
	Ledgers []store.ReadLedgerSummary `json:"ledgers"`
}

type accountListResponse struct {
	Accounts []store.ReadAccountDetail `json:"accounts"`
}

type assetListResponse struct {
	Assets []store.ReadAssetDetail `json:"assets"`
}

type contractListResponse struct {
	Contracts []store.ReadContractDetail `json:"contracts"`
}

type operationListResponse struct {
	Operations []store.ReadOperationSummary `json:"operations"`
}

type timelineResponse struct {
	Items []store.ReadTimelineItem `json:"items"`
}

type assetHolderListResponse struct {
	Holders []store.ReadAssetHolderSummary `json:"holders"`
}

type contractEventListResponse struct {
	Events []store.ReadContractEventSummary `json:"events"`
}

type contractStorageListResponse struct {
	Storage []store.ReadContractStorageSummary `json:"storage"`
}

type accountLookupResponse struct {
	Account            *store.ReadAccountDetail         `json:"account"`
	Trustlines         []store.ReadTrustlineSummary     `json:"trustlines"`
	Signers            []store.ReadAccountSignerSummary `json:"signers"`
	RecentTransactions []store.ReadTransactionSummary   `json:"recent_transactions"`
	RecentOperations   []store.ReadOperationSummary     `json:"recent_operations"`
}

type assetLookupResponse struct {
	Asset              *store.ReadAssetDetail         `json:"asset"`
	TopHolders         []store.ReadAssetHolderSummary `json:"top_holders"`
	RecentTransactions []store.ReadTransactionSummary `json:"recent_transactions"`
}

type contractLookupResponse struct {
	Contract           *store.ReadContractDetail          `json:"contract"`
	Spec               *store.ReadContractSpec            `json:"spec,omitempty"`
	Storage            []store.ReadContractStorageSummary `json:"storage"`
	RecentTransactions []store.ReadTransactionSummary     `json:"recent_transactions"`
	RecentEvents       []store.ReadContractEventSummary   `json:"recent_events"`
}

type searchResponse struct {
	Results []store.ReadSearchResult `json:"results"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, searchResponse{Results: []store.ReadSearchResult{}})
		return
	}

	limit := 10
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			writeErrorMessage(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed < limit {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	results, err := s.store.Search(ctx, query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if results == nil {
		results = []store.ReadSearchResult{}
	}

	writeJSON(w, http.StatusOK, searchResponse{Results: results})
}

func (s *Server) handleLedgerList(w http.ResponseWriter, r *http.Request) {
	before, err := optionalUint32Query(r, "before")
	if err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "before must be a positive ledger sequence")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ledgers, err := s.store.ListLedgerSummaries(ctx, lookupSliceLimit(r), before)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if ledgers == nil {
		ledgers = []store.ReadLedgerSummary{}
	}
	writeJSON(w, http.StatusOK, ledgerListResponse{Ledgers: ledgers})
}

func (s *Server) handleLedgerLookup(w http.ResponseWriter, r *http.Request) {
	value, child := splitLookupPath("/v1/ledgers/", r.URL.Path)
	if value == "" {
		writeErrorMessage(w, http.StatusBadRequest, "ledger sequence is required")
		return
	}

	sequence, err := strconv.ParseUint(value, 10, 32)
	if err != nil || sequence == 0 {
		writeErrorMessage(w, http.StatusBadRequest, "ledger sequence must be a positive integer")
		return
	}

	if child != "" && child != "transactions" {
		writeErrorMessage(w, http.StatusNotFound, "ledger subresource not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if child == "transactions" {
		transactions, err := s.store.ListTransactionSummariesByLedger(ctx, uint32(sequence), lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, transactionListResponse{Transactions: transactions})
		return
	}

	ledger, err := s.store.GetLedgerSummaryBySequence(ctx, uint32(sequence))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorMessage(w, http.StatusNotFound, "ledger not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	transactions, err := s.store.ListTransactionSummariesByLedger(ctx, uint32(sequence), lookupSliceLimit(r), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, ledgerLookupResponse{
		Ledger:       ledger,
		Transactions: transactions,
	})
}

func (s *Server) handleTransactionLookup(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimPrefix(r.URL.Path, "/v1/transactions/")
	hash = strings.TrimSpace(hash)
	if hash == "" {
		writeErrorMessage(w, http.StatusBadRequest, "transaction hash is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := s.store.GetTransactionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorMessage(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	ops, err := s.store.ListOperationsByTransactionHash(ctx, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.enrichReadOperations(ctx, ops); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	effects, err := s.store.ListEffectsByTransactionHash(ctx, hash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if effects == nil {
		effects = []store.ReadEffectSummary{}
	}

	writeJSON(w, http.StatusOK, transactionLookupResponse{
		Transaction: tx,
		Operations:  ops,
		Effects:     effects,
	})
}

func (s *Server) handleAccountList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	accounts, err := s.store.ListAccountDetails(ctx, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if accounts == nil {
		accounts = []store.ReadAccountDetail{}
	}
	writeJSON(w, http.StatusOK, accountListResponse{Accounts: accounts})
}

func (s *Server) handleAccountLookup(w http.ResponseWriter, r *http.Request) {
	id, child := splitLookupPath("/v1/accounts/", r.URL.Path)
	if id == "" {
		writeErrorMessage(w, http.StatusBadRequest, "account id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	switch child {
	case "":
	case "transactions":
		transactions, err := s.store.ListTransactionSummariesByAccount(ctx, id, lookupSliceLimit(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, transactionListResponse{Transactions: transactions})
		return
	case "operations":
		operations, err := s.store.ListOperationSummariesByAccount(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if err := s.enrichReadOperations(ctx, operations); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, operationListResponse{Operations: operations})
		return
	case "timeline":
		category, err := lookupTimelineCategory(r, "tx", "op")
		if err != nil {
			writeErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}
		items, err := s.store.ListAccountTimeline(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r), category)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if items == nil {
			items = []store.ReadTimelineItem{}
		}
		writeJSON(w, http.StatusOK, timelineResponse{Items: items})
		return
	default:
		writeErrorMessage(w, http.StatusNotFound, "account subresource not found")
		return
	}

	account, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorMessage(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	trustlines, err := s.store.ListTrustlinesByAccountID(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	signers, err := s.store.ListAccountSignersByAccountID(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentTransactions, err := s.store.ListTransactionSummariesByAccount(ctx, id, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentOperations, err := s.store.ListOperationSummariesByAccount(ctx, id, lookupSliceLimit(r), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.enrichReadOperations(ctx, recentOperations); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, accountLookupResponse{
		Account:            account,
		Trustlines:         trustlines,
		Signers:            signers,
		RecentTransactions: recentTransactions,
		RecentOperations:   recentOperations,
	})
}

func (s *Server) handleAssetList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	assets, err := s.store.ListAssetDetails(ctx, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if assets == nil {
		assets = []store.ReadAssetDetail{}
	}
	writeJSON(w, http.StatusOK, assetListResponse{Assets: assets})
}

func (s *Server) handleAssetLookup(w http.ResponseWriter, r *http.Request) {
	value, child := splitLookupPath("/v1/assets/", r.URL.Path)
	if value == "" {
		writeErrorMessage(w, http.StatusBadRequest, "asset code and issuer are required")
		return
	}
	code, issuer, ok := strings.Cut(value, ":")
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if !ok || code == "" || issuer == "" {
		writeErrorMessage(w, http.StatusBadRequest, "asset id must be CODE:ISSUER")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	switch child {
	case "":
	case "transactions":
		transactions, err := s.store.ListTransactionSummariesByAsset(ctx, code, issuer, lookupSliceLimit(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, transactionListResponse{Transactions: transactions})
		return
	case "holders":
		holders, err := s.store.ListAssetHoldersByCodeIssuer(ctx, code, issuer, lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, assetHolderListResponse{Holders: holders})
		return
	case "timeline":
		category, err := lookupTimelineCategory(r, "tx", "holder")
		if err != nil {
			writeErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}
		items, err := s.store.ListAssetTimeline(ctx, code, issuer, lookupSliceLimit(r), lookupSliceOffset(r), category)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if items == nil {
			items = []store.ReadTimelineItem{}
		}
		writeJSON(w, http.StatusOK, timelineResponse{Items: items})
		return
	default:
		writeErrorMessage(w, http.StatusNotFound, "asset subresource not found")
		return
	}

	asset, err := s.store.GetAssetByCodeIssuer(ctx, code, issuer)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorMessage(w, http.StatusNotFound, "asset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	topHolders, err := s.store.ListAssetHoldersByCodeIssuer(ctx, code, issuer, lookupSliceLimit(r), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentTransactions, err := s.store.ListTransactionSummariesByAsset(ctx, code, issuer, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, assetLookupResponse{
		Asset:              asset,
		TopHolders:         topHolders,
		RecentTransactions: recentTransactions,
	})
}

func (s *Server) handleContractList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	contracts, err := s.store.ListContractDetails(ctx, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if contracts == nil {
		contracts = []store.ReadContractDetail{}
	}
	writeJSON(w, http.StatusOK, contractListResponse{Contracts: contracts})
}

func (s *Server) handleContractLookup(w http.ResponseWriter, r *http.Request) {
	id, child := splitLookupPath("/v1/contracts/", r.URL.Path)
	if id == "" {
		writeErrorMessage(w, http.StatusBadRequest, "contract id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	switch child {
	case "":
	case "transactions":
		transactions, err := s.store.ListTransactionSummariesByContract(ctx, id, lookupSliceLimit(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, transactionListResponse{Transactions: transactions})
		return
	case "events":
		events, err := s.store.ListContractEventSummariesByContractID(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if err := s.enrichReadContractEvents(ctx, events); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, contractEventListResponse{Events: events})
		return
	case "spec":
		spec, err := s.store.GetContractSpecByID(ctx, id)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErrorMessage(w, http.StatusNotFound, "contract spec not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, spec)
		return
	case "storage":
		storage, err := s.store.ListContractStorageByContractID(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if storage == nil {
			storage = []store.ReadContractStorageSummary{}
		}
		writeJSON(w, http.StatusOK, contractStorageListResponse{Storage: storage})
		return
	case "invocations":
		operations, err := s.store.ListOperationSummariesByContract(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if operations == nil {
			operations = []store.ReadOperationSummary{}
		}
		if err := s.enrichReadOperations(ctx, operations); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, operationListResponse{Operations: operations})
		return
	case "timeline":
		category, err := lookupTimelineCategory(r, "tx", "event")
		if err != nil {
			writeErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}
		items, err := s.store.ListContractTimeline(ctx, id, lookupSliceLimit(r), lookupSliceOffset(r), category)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if items == nil {
			items = []store.ReadTimelineItem{}
		}
		writeJSON(w, http.StatusOK, timelineResponse{Items: items})
		return
	default:
		writeErrorMessage(w, http.StatusNotFound, "contract subresource not found")
		return
	}

	contract, err := s.store.GetContractByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorMessage(w, http.StatusNotFound, "contract not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentTransactions, err := s.store.ListTransactionSummariesByContract(ctx, id, lookupSliceLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recentEvents, err := s.store.ListContractEventSummariesByContractID(ctx, id, lookupSliceLimit(r), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.enrichReadContractEvents(ctx, recentEvents); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	spec, err := s.store.GetContractSpecByID(ctx, id)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	storage, err := s.store.ListContractStorageByContractID(ctx, id, lookupSliceLimit(r), 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if storage == nil {
		storage = []store.ReadContractStorageSummary{}
	}

	writeJSON(w, http.StatusOK, contractLookupResponse{
		Contract:           contract,
		Spec:               spec,
		Storage:            storage,
		RecentTransactions: recentTransactions,
		RecentEvents:       recentEvents,
	})
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeErrorMessage(w, status, err.Error())
}

func writeErrorMessage(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func splitLookupPath(prefix string, rawPath string) (string, string) {
	value := strings.TrimPrefix(rawPath, prefix)
	value = strings.Trim(value, "/")
	if value == "" {
		return "", ""
	}
	head, tail, found := strings.Cut(value, "/")
	if !found {
		return strings.TrimSpace(head), ""
	}
	return strings.TrimSpace(head), strings.TrimSpace(tail)
}

func lookupSliceLimit(r *http.Request) int {
	limit := 10
	value := strings.TrimSpace(r.URL.Query().Get("limit"))
	if value == "" {
		return limit
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return limit
	}
	if parsed > 50 {
		return 50
	}
	return parsed
}

func lookupSliceOffset(r *http.Request) int {
	value := strings.TrimSpace(r.URL.Query().Get("offset"))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func lookupTimelineCategory(r *http.Request, allowed ...string) (string, error) {
	value := strings.TrimSpace(r.URL.Query().Get("type"))
	if value == "" {
		value = strings.TrimSpace(r.URL.Query().Get("kind"))
	}
	if value == "" || strings.EqualFold(value, "all") || strings.EqualFold(value, "activity") {
		return "", nil
	}

	category, ok := normalizeTimelineCategory(value)
	if !ok {
		return "", errors.New("timeline type must be tx, op, holder, event, or all")
	}
	for _, candidate := range allowed {
		if category == candidate {
			return category, nil
		}
	}
	return "", errors.New("timeline type is not supported for this entity")
}

func normalizeTimelineCategory(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "tx", "transaction", "transactions":
		return "tx", true
	case "op", "ops", "operation", "operations":
		return "op", true
	case "holder", "holders":
		return "holder", true
	case "event", "events":
		return "event", true
	default:
		return "", false
	}
}

func optionalUint32Query(r *http.Request, name string) (uint32, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		if err != nil {
			return 0, err
		}
		return 0, errors.New("zero value")
	}
	return uint32(parsed), nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
