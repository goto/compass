package lineageparser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/goto/compass/core/asset"
)

const (
	columnLineageAPIPath     = "/api/v1/lineage/columns"
	columnLineageServiceType = "maxcompute"

	httpTimeout     = 30 * time.Second
	maxHTTPAttempts = 2
)

// HTTPClient implements asset.LineageParserClient by calling an external
// column-lineage HTTP service.
type HTTPClient struct {
	httpClient *http.Client
	host       string
}

func NewHTTPClient(host string) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{Timeout: httpTimeout},
		host:       host,
	}
}

// FetchColumnLineage calls the external lineage service with query and returns
// the parsed LineageGraph.  Returns nil, nil when the host is not configured.
func (c *HTTPClient) FetchColumnLineage(ctx context.Context, query string) (asset.LineageGraph, error) {
	if c.host == "" {
		return nil, nil
	}

	payload, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, fmt.Errorf("failed to encode column lineage payload: %w", err)
	}

	url := c.host + columnLineageAPIPath
	data, err := c.doRequest(ctx, url, payload)
	if err != nil {
		return nil, err
	}

	var response columnLineageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("parse column lineage response: %w", err)
	}

	return c.parseResponse(response), nil
}

func (c *HTTPClient) doRequest(ctx context.Context, url string, payload []byte) ([]byte, error) {
	var lastErr error
	for attempt := range maxHTTPAttempts {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create column lineage request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to send column lineage request: %w", attempt+1, err)
			continue
		}

		body, err := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: read column lineage response body: %w", attempt+1, err)
			continue
		}

		if res.StatusCode >= 300 {
			lastErr = fmt.Errorf("attempt %d: column lineage response status code %d: %s", attempt+1, res.StatusCode, string(body))
			continue
		}

		return body, nil
	}
	return nil, lastErr
}

func (c *HTTPClient) parseResponse(response columnLineageResponse) asset.LineageGraph {
	seen := make(map[string]struct{})
	var graph asset.LineageGraph

	if response.TargetTable == "" || len(strings.Split(response.TargetTable, ".")) != 3 {
		return graph
	}

	for _, col := range response.Columns {
		for _, src := range col.Sources {
			if src.Table == "" || len(strings.Split(src.Table, ".")) != 3 {
				continue
			}
			key := src.Table + "." + src.Column + "->" + response.TargetTable + "." + col.TargetColumn
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			graph = append(graph, asset.LineageEdge{
				Source:       c.toURN(src.Table),
				SourceColumn: src.Column,
				Target:       c.toURN(response.TargetTable),
				TargetColumn: col.TargetColumn,
			})
		}
	}

	return graph
}

// toURN converts a fully-qualified table name (e.g. "project.dataset.table")
// into a Compass URN using the configured service type.
func (*HTTPClient) toURN(fullTableName string) string {
	parts := strings.Split(fullTableName, ".")
	return fmt.Sprintf("urn:%s:%s:table:%s", columnLineageServiceType, parts[0], fullTableName)
}

type columnLineageResponse struct {
	TargetTable string                      `json:"target_table"`
	Columns     []columnLineageColumnDetail `json:"columns"`
}

type columnLineageColumnDetail struct {
	TargetColumn string                      `json:"target_column"`
	Sources      []columnLineageSourceDetail `json:"sources"`
}

type columnLineageSourceDetail struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}
