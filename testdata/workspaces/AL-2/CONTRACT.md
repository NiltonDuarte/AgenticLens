# Contract: AL-2 Go HTTP REST API and Embedded Filesystem Server

## 1. Problem Statement
Expose telemetry datasets via high-performance Go HTTP REST endpoints (/api/fleet, /api/workspaces, /api/workspaces/{ticketId}).

## 2. Invariants
- Query parameters window and sort must be strictly validated.
- Static files served from embed.FS fallback to index.html for SPA routes.
- Zero emojis in log lines or response payloads.
