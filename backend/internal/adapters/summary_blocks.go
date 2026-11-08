package adapters

import (
	"errors"
	"strconv"
	"strings"
)

const summaryMetricsStart = "<!-- agentic-lens:metrics:start -->"
const summaryMetricsEnd = "<!-- agentic-lens:metrics:end -->"

type summaryLine struct {
	start, end int
	text       string
	heading    int
	title      string
	opaque     bool
}

// summaryMarkdownLines recognizes headings/markers only outside fenced code.
func summaryMarkdownLines(content string) []summaryLine {
	var lines []summaryLine
	offset := 0
	fence := byte(0)
	fenceSize := 0
	for _, raw := range strings.SplitAfter(content, "\n") {
		if raw == "" {
			continue
		}
		line := summaryLine{start: offset, end: offset + len(raw), text: strings.TrimSuffix(strings.TrimSuffix(raw, "\n"), "\r")}
		offset += len(raw)
		text := strings.TrimLeft(line.text, " ")
		indent := len(line.text) - len(text)
		if indent > 3 {
			lines = append(lines, summaryLine{start: line.start, end: line.end, opaque: true})
			continue
		}
		run := 0
		if len(text) > 0 && (text[0] == '`' || text[0] == '~') {
			for run < len(text) && text[run] == text[0] {
				run++
			}
		}
		if fence != 0 {
			if run >= fenceSize && text[0] == fence && strings.TrimSpace(text[run:]) == "" {
				fence = 0
				fenceSize = 0
			}
			lines = append(lines, summaryLine{start: line.start, end: line.end, opaque: true})
			continue
		}
		if run >= 3 {
			fence = text[0]
			fenceSize = run
			lines = append(lines, summaryLine{start: line.start, end: line.end, opaque: true})
			continue
		}
		line.text = text
		level := 0
		for level < len(text) && text[level] == '#' {
			level++
		}
		if level >= 1 && level <= 6 && len(text) > level && (text[level] == ' ' || text[level] == '\t') {
			line.heading = level
			line.title = strings.TrimSpace(text[level:])
			// Markdown permits a closing hash run preceded by whitespace.
			trimmed := strings.TrimRight(line.title, "#")
			if len(trimmed) < len(line.title) && len(trimmed) > 0 && (trimmed[len(trimmed)-1] == ' ' || trimmed[len(trimmed)-1] == '\t') {
				line.title = strings.TrimSpace(trimmed)
			}
		}
		lines = append(lines, line)
	}
	for i := 0; i+1 < len(lines); i++ {
		line, next := &lines[i], lines[i+1]
		if line.opaque || next.opaque || line.heading != 0 || strings.TrimSpace(line.text) == "" {
			continue
		}
		underline := strings.TrimSpace(next.text)
		if underline == "" {
			continue
		}
		marker := underline[0]
		if marker != '=' && marker != '-' {
			continue
		}
		onlyMarker := true
		for j := range underline {
			if underline[j] != marker {
				onlyMarker = false
				break
			}
		}
		if onlyMarker {
			line.heading = 2
			if marker == '=' {
				line.heading = 1
			}
			line.title = line.text
		}
	}
	return lines
}

// legacySummaryEnd recognizes the formatter's reserved legacy vocabulary only.
// Unknown prose, code, headings and tables remain outside the replacement; a
// legacy section heading alone never grants ownership of all following lines.
func legacySummaryEnd(lines []summaryLine, index int) int {
	end := lines[index].end
	section := lines[index].title
	tableColumns := 0
	for _, line := range lines[index+1:] {
		if line.opaque {
			break
		}
		if strings.TrimSpace(line.text) == "" {
			continue
		}
		if line.heading != 0 {
			peer := line.heading == 3 && (line.title == "Anomalies & Health Flags" || line.title == "Agentic Execution Metrics")
			child := line.heading == 4 && section == "Agentic Execution Metrics" && (line.title == "By Subagent Role" || line.title == "By Model")
			if !peer && !child {
				break
			}
			if peer {
				section = line.title
			}
			tableColumns = 0
			end = line.end
			continue
		}
		if !legacySummaryContentOwned(line.text, section, &tableColumns) {
			break
		}
		end = line.end
	}
	return end
}

func legacySummaryContentOwned(text, section string, tableColumns *int) bool {
	if section == "Workflow Efficiency" {
		for _, label := range []string{
			"Iterations", "Software Development Life Cycle Iterations", "Active Agent Duration",
			"Context Window Health", "Context Window Usage", "Role Effort Distribution (Subagents)",
			"Subagent Turn Velocity & Thrashing", "Test-to-Production Code Ratio", "Code Delivered",
			"Reasoning Effort", "Tool Health & Self-Correction",
		} {
			if strings.HasPrefix(text, "- **"+label+":** ") {
				return true
			}
		}
	}
	if text == "Optional anomaly analysis: disabled by generation policy." || text == "Anomaly analysis is not recorded for every required owner/group; recorded findings remain visible. A complete source-backed refresh is required for recovery." {
		return true
	}
	if section == "Anomalies & Health Flags" {
		if text == "- **Status:** No execution anomalies detected (clean session)." || text == "- **Status:** Anomaly analysis is not recorded for the full ledger; no clean-session conclusion is available." {
			return true
		}
		if status, ok := strings.CutPrefix(text, "- **Status:** "); ok {
			count, rest, found := strings.Cut(status, " ")
			_, err := strconv.ParseUint(count, 10, 64)
			if found && err == nil && (strings.HasPrefix(rest, "execution anomalies detected (") || strings.HasPrefix(rest, "execution anomaly detected (")) && strings.HasSuffix(rest, ").") {
				return true
			}
		}
		for _, severity := range []string{"WARNING", "CRITICAL"} {
			for _, rule := range []string{"TOOL_RETRY_STREAK", "TOOL_HIGH_FAILURE_RATE", "CONTEXT_SATURATION_SPIKE", "SUBAGENT_REWORK_LOOP"} {
				if strings.HasPrefix(text, "- ["+severity+"] **"+rule+"**: ") {
					return true
				}
			}
		}
	}
	if section != "Agentic Execution Metrics" {
		return false
	}
	if strings.HasPrefix(text, "**Contributors:** ") {
		return true
	}
	for _, header := range []string{
		"| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |",
		"| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |",
	} {
		if text == header {
			*tableColumns = strings.Count(header, "|")
			return true
		}
	}
	if *tableColumns == 0 || strings.Count(text, "|") != *tableColumns {
		return false
	}
	if strings.Trim(text, "|-: ") == "" {
		return true
	}
	if !strings.HasPrefix(text, "| **") || !strings.HasSuffix(text, "|") {
		return false
	}
	fields := strings.Split(text, "|")
	count := strings.ReplaceAll(strings.Trim(fields[2], "* "), ",", "")
	_, err := strconv.ParseUint(count, 10, 64)
	return err == nil
}

func replaceSummaryMetrics(content, table string) (string, error) {
	lines := summaryMarkdownLines(content)
	start, end := -1, -1
	for _, line := range lines {
		if line.text == summaryMetricsStart {
			if start != -1 {
				return "", errors.New("ambiguous generated metrics start markers; repair markers without removing user evidence")
			}
			start = line.start
		}
		if line.text == summaryMetricsEnd {
			if start < 0 || end >= 0 {
				return "", errors.New("unmatched generated metrics end marker")
			}
			end = line.end
		}
	}
	if start >= 0 && end < 0 {
		return "", errors.New("unterminated generated metrics block; original summary preserved")
	}
	if start < 0 {
		for index, line := range lines {
			if line.heading != 3 || (line.title != "Workflow Efficiency" && line.title != "Agentic Execution Metrics") {
				continue
			}
			start = line.start
			end = legacySummaryEnd(lines, index)
			// An authored interruption can split the old generated sections.
			// Keeping a later stale health/table section beside the new block
			// would misrepresent the update; refuse without touching evidence.
			for _, remaining := range lines[index+1:] {
				if remaining.start < end {
					continue
				}
				peer := remaining.heading == 3 && (remaining.title == "Workflow Efficiency" || remaining.title == "Anomalies & Health Flags" || remaining.title == "Agentic Execution Metrics")
				child := remaining.heading == 4 && (remaining.title == "By Subagent Role" || remaining.title == "By Model")
				if peer || child {
					return "", errors.New("cannot safely replace noncontiguous legacy metrics around authored evidence; add metrics start/end markers around generated content only, keeping authored evidence outside those markers")
				}
			}
			break
		}
	}
	block := summaryMetricsStart + "\n" + strings.TrimSpace(table) + "\n" + summaryMetricsEnd + "\n"
	if start < 0 {
		if content == "" {
			return block, nil
		}
		separator := "\n\n"
		if strings.HasSuffix(content, "\n\n") {
			separator = ""
		} else if strings.HasSuffix(content, "\n") {
			separator = "\n"
		}
		return content + separator + block, nil
	}
	prefix, suffix := content[:start], content[end:]
	// Delimit generated content without trimming the user's surrounding bytes.
	if prefix != "" && !strings.HasSuffix(prefix, "\n") {
		prefix += "\n"
	}
	if suffix != "" && !strings.HasPrefix(suffix, "\n") {
		block += "\n"
	}
	return prefix + block + suffix, nil
}
