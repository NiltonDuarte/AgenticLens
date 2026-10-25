package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	adaptershttp "agenticlens/internal/adapters/http"
)

func TestCodexExtractHandlerPrunesDownloadedEventsAndPreservesTotals(t *testing.T) {
	fixture, err := os.ReadFile("../codex/testdata/rollout.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/codex/extract", strings.NewReader(string(fixture)))
	request.Header.Set("Content-Type", "application/x-ndjson")
	response := httptest.NewRecorder()
	adaptershttp.CodexExtractHandler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	var result codex.ExtractionResult
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Ledger.Facts.Events) != 1 || result.Ledger.Facts.Events[0].Status != "error" {
		t.Fatalf("download should retain only the fixture's error event: %#v", result.Ledger.Facts.Events)
	}
	if result.Transcript.ToolCalls != 2 || result.Ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 2 {
		t.Fatal("download pruning changed tool totals")
	}
	for _, marker := range []string{"/repo/example", "fixture-secret-command", "fixture-secret-args", "fixture-secret-prompt"} {
		if strings.Contains(response.Body.String(), marker) {
			t.Fatalf("extraction response leaked fixture content %q", marker)
		}
	}
}

func TestCodexExtractHandlerReturnsPrivacyMinimizedSummary(t *testing.T) {
	body := strings.Join([]string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"thread-1","session_id":"session-1","timestamp":"2026-10-25T09:00:00Z","cwd":"/tmp/project","source":"cli","cli_version":"test","model_provider":"openai"}}`,
		`{"timestamp":"2026-10-25T09:00:01Z","type":"token_usage_record","payload":{"usage":{"input_tokens":30,"output_tokens":6,"cached_input_tokens":12,"cache_write_input_tokens":2,"reasoning_output_tokens":1}}}`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"private-prompt-marker"}]}}`,
	}, "\n")
	request := httptest.NewRequest(http.MethodPost, "/api/codex/extract", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-ndjson")
	response := httptest.NewRecorder()

	adaptershttp.CodexExtractHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"transcript":{"thread_id":"thread-1"`) || !strings.Contains(response.Body.String(), `"ticket_id":"GEN-07"`) || !strings.Contains(response.Body.String(), `"schema_version":3`) {
		t.Fatalf("response is missing the Codex summary or AgenticLens ledger: %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "private-prompt-marker") {
		t.Fatal("response retained content-bearing user prompt text")
	}
	if strings.Contains(response.Body.String(), "/tmp/project") || strings.Contains(response.Body.String(), `"working_dir"`) {
		t.Fatal("response exposed the uploaded working-directory path")
	}
}

func TestCodexExtractHandlerAcceptsExplicitTicketAndRejectsInvalidTicketType(t *testing.T) {
	body := `{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"thread-1","timestamp":"2026-10-25T09:00:00Z"}}`
	request := httptest.NewRequest(http.MethodPost, "/api/codex/extract?ticket_id=GEN-07&ticket_title=Research&ticket_type=chore", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-ndjson")
	response := httptest.NewRecorder()
	adaptershttp.CodexExtractHandler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ticket_title":"Research"`) {
		t.Fatalf("expected explicit ticket metadata in generated ledger, got %d: %s", response.Code, response.Body.String())
	}

	invalidRequest := httptest.NewRequest(http.MethodPost, "/api/codex/extract?ticket_type=madeup", strings.NewReader(body))
	invalidRequest.Header.Set("Content-Type", "application/x-ndjson")
	invalidResponse := httptest.NewRecorder()
	adaptershttp.CodexExtractHandler().ServeHTTP(invalidResponse, invalidRequest)
	if invalidResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid ticket type to be rejected, got %d", invalidResponse.Code)
	}
}

func TestCodexExtractHandlerRejectsUnsupportedMediaType(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/codex/extract", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	adaptershttp.CodexExtractHandler().ServeHTTP(response, request)

	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status 415, got %d", response.Code)
	}
}

func TestCodexExtractHandlerRejectsFileWithoutValidRecords(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/codex/extract", strings.NewReader("not-jsonl"))
	request.Header.Set("Content-Type", "application/x-ndjson")
	response := httptest.NewRecorder()

	adaptershttp.CodexExtractHandler().ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", response.Code, response.Body.String())
	}
}
