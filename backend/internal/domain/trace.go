package domain

import "fmt"

// Validate checks the transient availability discriminator and native owner
// coverage. Committed facts never qualify as source trace evidence.
func (response WorkspaceEventsResponse) Validate() error {
	required := make(map[string]bool, len(response.Coverage.RequiredSessionIDs))
	for _, id := range response.Coverage.RequiredSessionIDs {
		if id == "" || required[id] {
			return fmt.Errorf("invalid required trace session %q", id)
		}
		required[id] = true
	}
	complete := make(map[string]bool, len(response.Coverage.CompleteSessionIDs))
	for _, id := range response.Coverage.CompleteSessionIDs {
		if !required[id] || complete[id] {
			return fmt.Errorf("invalid complete trace session %q", id)
		}
		complete[id] = true
	}
	for _, event := range response.Events {
		if !required[event.SessionID] {
			return fmt.Errorf("trace event has unrecorded owner %q", event.SessionID)
		}
	}
	switch response.Availability {
	case TraceComplete:
		if !response.IsHydrated || len(required) == 0 || len(complete) != len(required) || len(response.Coverage.Diagnostics) != 0 {
			return fmt.Errorf("complete trace requires successful complete native owner reads")
		}
	case TracePartial:
		if !response.IsHydrated || len(response.Events) == 0 && len(complete) == 0 {
			return fmt.Errorf("partial trace requires usable native evidence")
		}
		if len(complete) == len(required) && len(response.Coverage.Diagnostics) == 0 {
			return fmt.Errorf("partial trace requires incomplete coverage or source diagnostics")
		}
	case TraceUnavailable:
		if response.IsHydrated || len(response.Events) != 0 || len(complete) != 0 || response.DataSource != DataSourceUnavailable {
			return fmt.Errorf("unavailable trace cannot contain native evidence or a committed baseline")
		}
		return nil
	default:
		return fmt.Errorf("unsupported trace availability %q", response.Availability)
	}
	switch response.DataSource {
	case DataSourceLocalSQLite, DataSourceLocalCodex, DataSourceMixed:
		return nil
	default:
		return fmt.Errorf("hydrated trace requires native source provenance")
	}
}
