package lineageparser_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/lineageparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	c := lineageparser.NewHTTPClient("http://localhost:8080")
	assert.NotNil(t, c)
}

func TestHTTPClient_FetchColumnLineage_EmptyHost(t *testing.T) {
	c := lineageparser.NewHTTPClient("")
	graph, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
	assert.NoError(t, err)
	assert.Nil(t, graph)
}

func TestHTTPClient_FetchColumnLineage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/lineage/columns", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "SELECT 1", body["query"])
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"target_table": "project.dataset.table",
			"columns": []map[string]interface{}{
				{"target_column": "col_a", "sources": []map[string]string{{"table": "project.dataset.src", "column": "src_col"}}},
			},
		})
	}))
	defer srv.Close()
	c := lineageparser.NewHTTPClient(srv.URL)
	graph, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
	require.NoError(t, err)
	require.Len(t, graph, 1)
	assert.Equal(t, "urn:maxcompute:project:table:project.dataset.src", graph[0].Source)
	assert.Equal(t, "src_col", graph[0].SourceColumn)
	assert.Equal(t, "urn:maxcompute:project:table:project.dataset.table", graph[0].Target)
	assert.Equal(t, "col_a", graph[0].TargetColumn)
}

func TestHTTPClient_FetchColumnLineage_RetriesOnServerError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()
	c := lineageparser.NewHTTPClient(srv.URL)
	_, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Equal(t, 2, attempts, "expected exactly 2 attempts")
	assert.Contains(t, err.Error(), "500")
}

func TestHTTPClient_FetchColumnLineage_ContextCancelled(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-block
		w.WriteHeader(200)
	}))
	defer srv.Close()
	defer close(block)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := lineageparser.NewHTTPClient(srv.URL)
	_, err := c.FetchColumnLineage(ctx, "SELECT 1")
	assert.Error(t, err)
}

func TestHTTPClient_FetchColumnLineage_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()
	c := lineageparser.NewHTTPClient(srv.URL)
	_, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
	assert.ErrorContains(t, err, "parse column lineage response")
}

func TestHTTPClient_FetchColumnLineage_ParseResponse(t *testing.T) {
	tests := []struct {
		name      string
		response  map[string]interface{}
		wantEdges int
	}{
		{
			name:      "empty target_table produces empty graph",
			response:  map[string]interface{}{"target_table": "", "columns": []interface{}{}},
			wantEdges: 0,
		},
		{
			name:      "target_table with wrong segment count produces empty graph",
			response:  map[string]interface{}{"target_table": "only.two", "columns": []interface{}{}},
			wantEdges: 0,
		},
		{
			name: "source table with wrong segment count is skipped",
			response: map[string]interface{}{
				"target_table": "p.d.t",
				"columns": []map[string]interface{}{
					{"target_column": "c", "sources": []map[string]string{{"table": "bad.src", "column": "c"}}},
				},
			},
			wantEdges: 0,
		},
		{
			name: "duplicate edges are deduplicated",
			response: map[string]interface{}{
				"target_table": "p.d.t",
				"columns": []map[string]interface{}{
					{"target_column": "c", "sources": []map[string]string{{"table": "p.d.src", "column": "c"}, {"table": "p.d.src", "column": "c"}}},
				},
			},
			wantEdges: 1,
		},
		{
			name: "multiple distinct edges are all included",
			response: map[string]interface{}{
				"target_table": "p.d.t",
				"columns": []map[string]interface{}{
					{"target_column": "col_a", "sources": []map[string]string{{"table": "p.d.src1", "column": "c1"}}},
					{"target_column": "col_b", "sources": []map[string]string{{"table": "p.d.src2", "column": "c2"}}},
				},
			},
			wantEdges: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tc.response)
			}))
			defer srv.Close()
			c := lineageparser.NewHTTPClient(srv.URL)
			graph, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
			require.NoError(t, err)
			assert.Len(t, graph, tc.wantEdges)
		})
	}
}

func TestHTTPClient_FetchColumnLineage_URNFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"target_table": "project.dataset.target_table",
			"columns": []map[string]interface{}{
				{"target_column": "col", "sources": []map[string]string{{"table": "project.dataset.source_table", "column": "source_col"}}},
			},
		})
	}))
	defer srv.Close()
	c := lineageparser.NewHTTPClient(srv.URL)
	graph, err := c.FetchColumnLineage(context.Background(), "SELECT 1")
	require.NoError(t, err)
	require.Len(t, graph, 1)
	assert.Equal(t, asset.LineageEdge{
		Source:       "urn:maxcompute:project:table:project.dataset.source_table",
		SourceColumn: "source_col",
		Target:       "urn:maxcompute:project:table:project.dataset.target_table",
		TargetColumn: "col",
	}, graph[0])
}
