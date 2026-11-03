package adapters

import (
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// formatValue retains ordered members and original scalar lexemes. It is a JSON
// syntax tree, deliberately independent of domain facts and numeric conversion.
type formatValue struct {
	kind     byte
	raw      []byte
	members  []formatMember
	elements []*formatValue
}
type formatMember struct {
	key   string
	raw   []byte
	value *formatValue
}
type formatParser struct {
	data []byte
	pos  int
}

func (p *formatParser) space() {
	for p.pos < len(p.data) && strings.ContainsRune(" \t\r\n", rune(p.data[p.pos])) {
		p.pos++
	}
}
func (p *formatParser) quoted() []byte {
	start := p.pos
	p.pos++
	for p.data[p.pos] != '"' {
		if p.data[p.pos] == '\\' {
			p.pos++
		}
		p.pos++
	}
	p.pos++
	return p.data[start:p.pos]
}
func (p *formatParser) value() (*formatValue, error) {
	p.space()
	n := &formatValue{kind: p.data[p.pos]}
	switch n.kind {
	case '{':
		p.pos++
		p.space()
		seen := map[string]bool{}
		for p.data[p.pos] != '}' {
			raw := p.quoted()
			var key string
			if err := json.Unmarshal(raw, &key); err != nil {
				return nil, err
			}
			if seen[key] {
				return nil, fmt.Errorf("duplicate JSON member %q", key)
			}
			seen[key] = true
			p.space()
			p.pos++
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			n.members = append(n.members, formatMember{key, raw, v})
			p.space()
			if p.data[p.pos] != ',' {
				break
			}
			p.pos++
			p.space()
		}
		p.pos++
	case '[':
		p.pos++
		p.space()
		for p.data[p.pos] != ']' {
			v, err := p.value()
			if err != nil {
				return nil, err
			}
			n.elements = append(n.elements, v)
			p.space()
			if p.data[p.pos] != ',' {
				break
			}
			p.pos++
			p.space()
		}
		p.pos++
	case '"':
		n.raw = p.quoted()
	default:
		start := p.pos
		for p.pos < len(p.data) && !strings.ContainsRune(",]} \r\n\t", rune(p.data[p.pos])) {
			p.pos++
		}
		n.raw = p.data[start:p.pos]
	}
	return n, nil
}
func (n *formatValue) member(key string) *formatValue {
	for _, m := range n.members {
		if m.key == key {
			return m.value
		}
	}
	return nil
}
func nullableKind(n *formatValue, kind byte) bool {
	return n == nil || n.kind == kind || bytes.Equal(n.raw, []byte("null"))
}
func validateFormatShape(n *formatValue) error {
	if n.kind != '{' || n.member("ticket_id") == nil || !nullableKind(n.member("ticket_id"), '"') {
		return errors.New("unsupported ledger shape: expected ledger object with ticket_id")
	}
	if v := n.member("schema_version"); v != nil && !bytes.Equal(v.raw, []byte("null")) {
		var version int
		if json.Unmarshal(v.raw, &version) != nil || version < 0 || version > domain.CurrentSchemaVersion {
			return errors.New("unsupported ledger schema_version")
		}
	}
	if n.member("sessions") == nil && n.member("invocations") == nil && n.member("totals") == nil {
		return errors.New("unsupported ledger shape: expected sessions, legacy invocations, or computed totals")
	}
	return validateFormatCollections(n, true)
}
func validateFormatCollections(n *formatValue, root bool) error {
	for _, key := range []string{"sessions", "invocations", "tools", "anomalies"} {
		if !root && key == "sessions" {
			continue
		}
		v := n.member(key)
		if !nullableKind(v, '[') {
			return fmt.Errorf("unsupported ledger shape: %s must be array or null", key)
		}
		if v == nil {
			continue
		}
		for _, item := range v.elements {
			if item.kind != '{' {
				return fmt.Errorf("unsupported ledger shape: %s entries must be objects", key)
			}
			if key == "sessions" {
				if err := validateFormatCollections(item, false); err != nil {
					return err
				}
			}
			if key == "invocations" {
				for _, field := range []string{"tokens", "cost_evidence", "evidence"} {
					if !nullableKind(item.member(field), '{') {
						return fmt.Errorf("unsupported invocation %s: expected object or null", field)
					}
				}
			}
		}
	}
	return nil
}
func compactFormatPath(path string) bool {
	switch path {
	case "tools/[]", "anomalies/[]", "sessions/[]/tools/[]", "sessions/[]/anomalies/[]",
		"invocations/[]/tokens", "invocations/[]/cost_evidence", "invocations/[]/evidence",
		"sessions/[]/invocations/[]/tokens", "sessions/[]/invocations/[]/cost_evidence", "sessions/[]/invocations/[]/evidence":
		return true
	}
	return false
}
func renderFormatValue(out *bytes.Buffer, n *formatValue, path string, depth int, compact bool) {
	compact = compact || compactFormatPath(path)
	if n.kind != '{' && n.kind != '[' {
		out.Write(n.raw)
		return
	}
	open, close := byte('{'), byte('}')
	count := len(n.members)
	if n.kind == '[' {
		open, close = '[', ']'
		count = len(n.elements)
	}
	out.WriteByte(open)
	for i := 0; i < count; i++ {
		if i > 0 {
			out.WriteByte(',')
			if compact {
				out.WriteByte(' ')
			}
		}
		if !compact {
			out.WriteByte('\n')
			out.WriteString(strings.Repeat("  ", depth+1))
		}
		childPath := path
		if childPath != "" {
			childPath += "/"
		}
		if n.kind == '{' {
			m := n.members[i]
			out.Write(m.raw)
			out.WriteString(": ")
			renderFormatValue(out, m.value, childPath+strings.ReplaceAll(strings.ReplaceAll(m.key, "~", "~0"), "/", "~1"), depth+1, compact)
		} else {
			renderFormatValue(out, n.elements[i], childPath+"[]", depth+1, compact)
		}
	}
	if count > 0 && !compact {
		out.WriteByte('\n')
		out.WriteString(strings.Repeat("  ", depth))
	}
	out.WriteByte(close)
}

// FormatLedgerJSON applies the installed release's whitespace policy without
// decoding facts, changing scalar lexemes, or normalizing unknown members.
func FormatLedgerJSON(data []byte) ([]byte, error) {
	if !json.Valid(data) {
		return nil, errors.New("invalid ledger JSON")
	}
	p := formatParser{data: data}
	n, err := p.value()
	if err != nil {
		return nil, err
	}
	if err = validateFormatShape(n); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	renderFormatValue(&out, n, "", 0, false)
	out.WriteByte('\n')
	return out.Bytes(), nil
}
