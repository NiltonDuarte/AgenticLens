package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

const sandboxTable = "sandbox_workspace_write"
const rootsKey = "writable_roots"

func tomlPathLiteral(path string) string {
	// JSON string escapes produced for valid filesystem UTF-8 are valid TOML
	// basic-string escapes, including \uXXXX for control characters.
	encoded, _ := json.Marshal(path)
	return string(encoded)
}

func initialInboxConfig(root string) string {
	return sandboxTable + "." + rootsKey + " = [" + tomlPathLiteral(root) + "]\n"
}

func expressionKey(n *unstable.Node) []string {
	var keys []string
	for key := n.Key(); key.Next(); {
		keys = append(keys, string(key.Node().Data))
	}
	return keys
}

func replaceRange(data []byte, from, to int, replacement string) []byte {
	result := make([]byte, 0, len(data)-to+from+len(replacement))
	result = append(result, data[:from]...)
	result = append(result, replacement...)
	return append(result, data[to:]...)
}

func end(raw unstable.Range) int { return int(raw.Offset + raw.Length) }

func valueStart(data []byte, kv *unstable.Node) int {
	lastKey := 0
	for k := kv.Key(); k.Next(); {
		lastKey = end(k.Node().Raw)
	}
	start := lastKey + bytes.IndexByte(data[lastKey:], '=') + 1
	for data[start] == ' ' || data[start] == '\t' {
		start++
	}
	return start
}

// TOML keys may be dotted, quoted or inline. Parse their identities and patch
// only the relevant token spans so comments, unrelated tables and ordering survive.
func editWritableRoot(data []byte, root string, remove bool) ([]byte, bool, error) {
	var schema struct {
		Sandbox struct {
			Roots []string `toml:"writable_roots"`
		} `toml:"sandbox_workspace_write"`
	}
	if err := toml.Unmarshal(data, &schema); err != nil {
		return nil, false, err
	}
	after, changed, err := patchWritableRoot(data, root, remove)
	if err != nil || !changed {
		return after, changed, err
	}
	expected := slices.Clone(schema.Sandbox.Roots)
	if remove {
		expected = slices.DeleteFunc(expected, func(s string) bool { return s == root })
	} else {
		expected = append([]string{root}, expected...)
	}
	schema.Sandbox.Roots = nil
	if err := toml.Unmarshal(after, &schema); err != nil {
		return nil, false, err
	}
	if !slices.Equal(schema.Sandbox.Roots, expected) {
		return nil, false, errors.New("edited configuration did not preserve writable root choices")
	}
	return after, true, nil
}

func patchWritableRoot(data []byte, root string, remove bool) ([]byte, bool, error) {
	var p unstable.Parser
	p.Reset(data)
	var section []string
	at := 0
	field := initialInboxConfig(root)
	for p.NextExpression() {
		n := p.Expression()
		if n.Kind == unstable.Table || n.Kind == unstable.ArrayTable {
			section = expressionKey(n)
			if slices.Equal(section, []string{sandboxTable}) {
				if n.Kind != unstable.Table {
					return nil, false, errors.New("sandbox workspace settings must be a table")
				}
				pos := end(n.Child().Raw)
				lineEnd := bytes.IndexByte(data[pos:], '\n')
				field = rootsKey + " = [" + tomlPathLiteral(root) + "]\n"
				if lineEnd == -1 {
					at = len(data)
					field = "\n" + field
				} else {
					at = pos + lineEnd + 1
				}
			}
			continue
		}
		if n.Kind != unstable.KeyValue {
			continue
		}
		path := slices.Concat(section, expressionKey(n))
		if slices.Equal(path, []string{sandboxTable}) {
			if n.Value().Kind != unstable.InlineTable {
				return nil, false, errors.New("sandbox workspace settings must be a table")
			}
			at = valueStart(data, n) + 1
			field = rootsKey + " = [" + tomlPathLiteral(root) + "]"
			if n.Value().Child() != nil {
				field += ", "
			}
			for child := n.Value().Children(); child.Next(); {
				if slices.Equal(expressionKey(child.Node()), []string{rootsKey}) {
					return editRootArray(data, child.Node(), root, remove)
				}
			}
		}
		if slices.Equal(path, []string{sandboxTable, rootsKey}) {
			return editRootArray(data, n, root, remove)
		}
	}
	if err := p.Error(); err != nil {
		return nil, false, err
	}
	if remove {
		return data, false, nil
	}
	return replaceRange(data, at, at, field), true, nil
}

func editRootArray(data []byte, kv *unstable.Node, root string, remove bool) ([]byte, bool, error) {
	a := kv.Value()
	if a.Kind != unstable.Array {
		return nil, false, errors.New("writable_roots must be a string array")
	}
	var hit unstable.Range
	for items := a.Children(); items.Next(); {
		n := items.Node()
		if n.Kind != unstable.String {
			return nil, false, errors.New("writable_roots contains a non-string value")
		}
		if string(n.Data) == root {
			if hit.Length != 0 && remove {
				return nil, false, errors.New("owned inbox grant is duplicated; resolve before removing")
			}
			hit = n.Raw
		}
	}
	if !remove {
		if hit.Length != 0 {
			return data, false, nil
		}
		start := valueStart(data, kv) + 1
		token := tomlPathLiteral(root)
		if a.Child() != nil {
			token += ", "
		}
		return replaceRange(data, start, start, token), true, nil
	}
	if hit.Length == 0 {
		return data, false, nil
	}
	// Remove the token and its following comma, if present. TOML permits a
	// trailing comma after a previous item. Comments on either side remain.
	tail := end(hit)
	for i := tail; i < end(kv.Raw); i++ {
		switch data[i] {
		case ' ', '\t', '\r', '\n':
		case '#':
			lineEnd := bytes.IndexByte(data[i:], '\n')
			if lineEnd < 0 {
				return nil, false, errors.New("unterminated array comment")
			}
			i += lineEnd
		case ',':
			after := i + 1
			if i == tail && after < len(data) && data[after] == ' ' {
				after++
			}
			result := replaceRange(data, i, after, "")
			return replaceRange(result, int(hit.Offset), tail, ""), true, nil
		default:
			return replaceRange(data, int(hit.Offset), tail, ""), true, nil
		}
	}
	return nil, false, errors.New("unterminated writable_roots array")
}
