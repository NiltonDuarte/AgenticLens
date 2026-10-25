package http

import (
	"errors"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

const maxCodexUploadBytes = 64 << 20

var codexTicketIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

// CodexExtractHandler parses one user-selected rollout JSONL in memory and
// returns a privacy-minimized summary plus the derived AgenticLens ledger.
func CodexExtractHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || (mediaType != "application/x-ndjson" && mediaType != "application/jsonl") {
			WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "upload a Codex rollout JSONL file")
			return
		}
		ticketID := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ticket_id")))
		if ticketID == "" {
			ticketID = "AL-56"
		}
		if !codexTicketIDPattern.MatchString(ticketID) {
			WriteError(w, http.StatusBadRequest, "invalid_ticket_id", "ticket_id must start with a letter and contain only letters, numbers, hyphens, or underscores")
			return
		}
		ticketTypeRaw := strings.TrimSpace(r.URL.Query().Get("ticket_type"))
		if ticketTypeRaw == "" {
			ticketTypeRaw = "chore"
		}
		ticketType, validTicketType := domain.NormalizeTicketType(ticketTypeRaw)
		if !validTicketType {
			WriteError(w, http.StatusBadRequest, "invalid_ticket_type", "ticket_type must be a conventional commit type")
			return
		}

		body := http.MaxBytesReader(w, r.Body, maxCodexUploadBytes)
		transcript, err := codex.ParseRollout(body)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				WriteError(w, http.StatusRequestEntityTooLarge, "upload_too_large", "Codex rollout files must be 64 MiB or smaller")
				return
			}
			if strings.Contains(err.Error(), "token too long") {
				WriteError(w, http.StatusRequestEntityTooLarge, "record_too_large", "a Codex rollout record exceeds the parser's 8 MiB line limit")
				return
			}
			WriteError(w, http.StatusBadRequest, "invalid_rollout", "could not read the Codex rollout JSONL")
			return
		}
		if len(transcript.RecordKinds) == 0 {
			WriteError(w, http.StatusUnprocessableEntity, "no_valid_records", "the selected file contains no valid Codex rollout records")
			return
		}
		ledger := transcript.ToMetricsLedger(ticketID, strings.TrimSpace(r.URL.Query().Get("ticket_title")), ticketType, domain.DefaultProjectConfig())
		WriteJSON(w, http.StatusOK, codex.ExtractionResult{Transcript: transcript, Ledger: ledger})
	}
}
