import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { ArtifactViewer } from './ArtifactViewer';

declare global {
  interface Window {
    __injected_script__?: boolean;
    __injected_img__?: boolean;
    __injected_svg__?: boolean;
  }
}

describe('ArtifactViewer Component', () => {
  const sampleContract = `# Contract: AL-22 Rich Markdown
Here is the contract specification for **AgenticLens**.

## Invariants
- Zero emojis permitted
- Strict typing

> Pure domain logic must remain isolated.

### Code Example
Inline \`formatTokens(1000)\` badge.

\`\`\`ts
export function formatCurrency(amount: number): string {
  return '$' + amount.toFixed(2);
}
\`\`\`

### Roadmap
- [x] Install Markdown dependencies
- [ ] Implement custom renderers

### Telemetry Summary
| Metric | Floor | Ceiling |
| :--- | :--- | :--- |
| Invocations | 1 | 50 |
| Duration | 120ms | 4500ms |
`;

  const sampleSummary = `# Summary: AL-22 Delivery
- All tasks delivered successfully.
- Code coverage at 100%.
`;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders rich document headings, paragraphs, blockquotes, and lists by default', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    // Markdown container exists
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();

    // Headings
    expect(screen.getByRole('heading', { level: 1, name: 'Contract: AL-22 Rich Markdown' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Invariants' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 3, name: 'Code Example' })).toBeInTheDocument();

    // Paragraph
    expect(screen.getByText(/Here is the contract specification for/)).toBeInTheDocument();

    // Blockquote
    const quote = document.querySelector('.markdown-blockquote');
    expect(quote).toBeInTheDocument();
    expect(quote).toHaveTextContent('Pure domain logic must remain isolated.');

    // List items
    expect(screen.getByText('Zero emojis permitted')).toBeInTheDocument();
    expect(screen.getByText('Strict typing')).toBeInTheDocument();
  });

  it('renders GitHub Flavored Markdown tables with styled headers, cells, and responsive wrapper', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const tableWrapper = document.querySelector('.markdown-table-wrapper');
    expect(tableWrapper).toBeInTheDocument();

    const table = tableWrapper?.querySelector('table.markdown-table');
    expect(table).toBeInTheDocument();

    const headers = screen.getAllByRole('columnheader');
    const headerTexts = headers.map((th) => th.textContent);
    expect(headerTexts).toContain('Metric');
    expect(headerTexts).toContain('Floor');
    expect(headerTexts).toContain('Ceiling');

    expect(screen.getByRole('cell', { name: 'Invocations' })).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: '1' })).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: '50' })).toBeInTheDocument();
  });

  it('renders task lists with checkboxes reflecting checked and unchecked states', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[];
    expect(checkboxes.length).toBe(2);

    // First task is checked
    expect(checkboxes[0]).toBeChecked();
    expect(checkboxes[0]).toBeDisabled();
    expect(checkboxes[0]).toHaveClass('markdown-checkbox');

    // Second task is unchecked
    expect(checkboxes[1]).not.toBeChecked();
    expect(checkboxes[1]).toBeDisabled();
    expect(checkboxes[1]).toHaveClass('markdown-checkbox');
  });

  it('renders inline code badges and formatted code blocks', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    // Inline code
    const inlineCode = document.querySelector('.markdown-code-inline');
    expect(inlineCode).toBeInTheDocument();
    expect(inlineCode).toHaveTextContent('formatTokens(1000)');

    // Pre code block
    const preBlock = document.querySelector('pre.markdown-pre');
    expect(preBlock).toBeInTheDocument();
    expect(preBlock).toHaveTextContent('export function formatCurrency');
  });

  it('toggles between rich document and raw source views', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    // Default: Rich document mode
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).not.toBeInTheDocument();

    const rawBtn = screen.getByRole('button', { name: 'Raw Source' });
    const richBtn = screen.getByRole('button', { name: 'Rich Document' });

    // Switch to Raw Source
    fireEvent.click(rawBtn);
    expect(document.querySelector('.artifact-markdown-container')).not.toBeInTheDocument();
    const pre = document.querySelector('.artifact-pre');
    expect(pre).toBeInTheDocument();
    expect(pre).toHaveTextContent('# Contract: AL-22 Rich Markdown');

    // Switch back to Rich Document
    fireEvent.click(richBtn);
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).not.toBeInTheDocument();
  });

  it('switches between CONTRACT.md and SUMMARY.md documents', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    expect(screen.getByText(/Contract: AL-22 Rich Markdown/)).toBeInTheDocument();
    expect(screen.queryByText(/Summary: AL-22 Delivery/)).not.toBeInTheDocument();

    const summaryTab = screen.getByRole('button', { name: /summary\.md/i });
    fireEvent.click(summaryTab);

    expect(screen.getByText(/Summary: AL-22 Delivery/)).toBeInTheDocument();
    expect(screen.queryByText(/Contract: AL-22 Rich Markdown/)).not.toBeInTheDocument();

    const contractTab = screen.getByRole('button', { name: /contract\.md/i });
    fireEvent.click(contractTab);

    expect(screen.getByText(/Contract: AL-22 Rich Markdown/)).toBeInTheDocument();
  });

  it('copies content to clipboard and displays Copied state for 2 seconds', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    expect(copyBtn).toBeInTheDocument();
    expect(copyBtn).not.toBeDisabled();

    // Trigger copy
    await act(async () => {
      fireEvent.click(copyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(1);
    expect(writeTextMock).toHaveBeenCalledWith(sampleContract);
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Fast-forward 1999ms: should still say Copied
    act(() => {
      vi.advanceTimersByTime(1999);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Fast-forward past 2000ms: should revert to Copy
    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(screen.getByRole('button', { name: 'Copy' })).toBeInTheDocument();
  });

  it('resets copy state when switching between artifact tabs', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    await act(async () => {
      fireEvent.click(copyBtn);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Switching to summary tab resets copy state to Copy immediately
    const summaryTab = screen.getByRole('button', { name: /summary\.md/i });
    fireEvent.click(summaryTab);
    expect(screen.getByRole('button', { name: 'Copy' })).toBeInTheDocument();
  });

  it('executes fallback copy mechanism when navigator.clipboard is unavailable', async () => {
    // Delete navigator.clipboard
    const originalClipboard = navigator.clipboard;
    // @ts-expect-error test fallback branch
    delete navigator.clipboard;

    const execCommandMock = vi.fn().mockReturnValue(true);
    document.execCommand = execCommandMock;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: 'Fallback text test',
          summary_content: null,
        }}
      />
    );

    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    await act(async () => {
      fireEvent.click(copyBtn);
    });

    expect(execCommandMock).toHaveBeenCalledWith('copy');
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Restore clipboard
    Object.assign(navigator, { clipboard: originalClipboard });
  });

  it('defensively handles null, undefined, empty, or whitespace contents', () => {
    const { rerender } = render(<ArtifactViewer artifacts={null} />);
    expect(screen.getByText('No CONTRACT.md document found for this workspace.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Copy' })).toBeDisabled();

    rerender(<ArtifactViewer artifacts={undefined} />);
    expect(screen.getByText('No CONTRACT.md document found for this workspace.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Copy' })).toBeDisabled();

    rerender(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: '   \n\t  ',
          summary_content: '',
        }}
      />
    );
    expect(screen.getByText('No CONTRACT.md document found for this workspace.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Copy' })).toBeDisabled();

    // Tab button shows (Empty) suffix
    expect(screen.getByRole('button', { name: /contract\.md \(empty\)/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /summary\.md \(empty\)/i })).toBeInTheDocument();

    // Clicking copy while disabled does nothing
    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    fireEvent.click(copyBtn);
    expect(copyBtn).toHaveTextContent('Copy');
  });

  it('gracefully renders malformed, incomplete, and unclosed markdown structures without crashing', () => {
    const malformedContent = `# Malformed Document Testing

### Table With Dangling Pipes and Uneven Cells
| Metric | Floor | Ceiling |
| :--- | :--- | :--- |
| Valid Row | 10 | 20 |
| Dangling Pipes Row | 30 | 40 | Extra trailing cell |
| Missing Cell Row | 50 |
| | | |

### Non-Standard Broken Table Delimiters
| Not | A | Real | Table |
| --- |
Raw text row without table delimiters

### Broken Brackets and Links
[Unclosed link bracket(https://example.com)
![Unclosed image alt text(https://example.com/banner.png
[Broken reference][missing-ref-tag]
[[Nested [broken] brackets]]

### Broken Formatting
> Unclosed blockquote with *unclosed italic and **unclosed bold
> > Double nested blockquote without terminating newline

### Unclosed Inline Code
Here is an unclosed \`inline code segment without closing backtick.

### Unclosed Code Fence at Document End
\`\`\`typescript
function unclosedBlockAtEOF() {
  const dangling = true;
`;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: malformedContent,
          summary_content: null,
        }}
      />
    );

    const container = document.querySelector('.artifact-markdown-container');
    expect(container).toBeInTheDocument();

    expect(screen.getByRole('heading', { level: 1, name: 'Malformed Document Testing' })).toBeInTheDocument();

    // Table with uneven/dangling cells still safely renders as a table element
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: 'Dangling Pipes Row' })).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: 'Missing Cell Row' })).toBeInTheDocument();

    // Non-standard broken table delimiter safely falls back to standard text
    expect(screen.getByText(/Raw text row without table delimiters/)).toBeInTheDocument();

    // Verify unclosed code fence at EOF renders inside a pre/code block without throwing
    expect(screen.getByText(/function unclosedBlockAtEOF/)).toBeInTheDocument();

    // Verify blockquote rendered
    const quote = document.querySelector('.markdown-blockquote');
    expect(quote).toBeInTheDocument();
    expect(quote).toHaveTextContent(/Unclosed blockquote with/);

    // Verify broken bracket text is safely displayed as literal text
    expect(screen.getByText(/\[Unclosed link bracket/)).toBeInTheDocument();
  });

  it('safely handles HTML injection and XSS payloads without executing scripts or injecting raw script tags', () => {
    const xssPayload = `# Security Injection Test
<script>window.__injected_script__ = true;</script>
<img src="nonexistent-target.png" onerror="window.__injected_img__ = true;" />
<iframe src="javascript:alert('xss-iframe')"></iframe>
<svg onload="window.__injected_svg__ = true"></svg>
<object data="javascript:alert('xss-object')"></object>
[Exploit Link](javascript:alert('xss-link'))
<a href="javascript:alert('raw-a-tag')">Raw Link Injection</a>
`;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: xssPayload,
          summary_content: null,
        }}
      />
    );

    const container = document.querySelector('.artifact-markdown-container');
    expect(container).toBeInTheDocument();

    // Verify dangerous elements are not rendered in DOM
    expect(container?.querySelector('script')).toBeNull();
    expect(container?.querySelector('iframe')).toBeNull();
    expect(container?.querySelector('object')).toBeNull();
    expect(container?.querySelector('svg')).toBeNull();

    // Verify image tags do not execute or possess onerror attributes
    const imgElements = container?.querySelectorAll('img') ?? [];
    imgElements.forEach((img) => {
      expect(img.getAttribute('onerror')).toBeNull();
    });

    // Verify all rendered links sanitize or omit javascript: URLs
    const links = container?.querySelectorAll('a') ?? [];
    links.forEach((link) => {
      const href = link.getAttribute('href') ?? '';
      expect(href.toLowerCase()).not.toContain('javascript:');
    });

    // Window globals remain unset
    expect(window.__injected_script__).toBeUndefined();
    expect(window.__injected_img__).toBeUndefined();
    expect(window.__injected_svg__).toBeUndefined();
  });

  it('handles rapid consecutive view mode and tab toggling without state corruption or display drift', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const rawBtn = screen.getByRole('button', { name: 'Raw Source' });
    const richBtn = screen.getByRole('button', { name: 'Rich Document' });
    const contractTab = screen.getByRole('button', { name: /^contract\.md/i });
    const summaryTab = screen.getByRole('button', { name: /^summary\.md/i });

    // Rapidly toggle view mode 10 times consecutively
    for (let i = 0; i < 10; i++) {
      fireEvent.click(rawBtn);
      expect(document.querySelector('.artifact-pre')).toBeInTheDocument();
      expect(document.querySelector('.artifact-markdown-container')).not.toBeInTheDocument();
      expect(rawBtn).toHaveClass('active');
      expect(richBtn).not.toHaveClass('active');

      fireEvent.click(richBtn);
      expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
      expect(document.querySelector('.artifact-pre')).not.toBeInTheDocument();
      expect(richBtn).toHaveClass('active');
      expect(rawBtn).not.toHaveClass('active');
    }

    // Toggle view mode to raw, then switch tabs repeatedly
    fireEvent.click(rawBtn);
    expect(document.querySelector('.artifact-pre')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).toHaveTextContent('Contract: AL-22 Rich Markdown');

    fireEvent.click(summaryTab);
    expect(document.querySelector('.artifact-pre')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).toHaveTextContent('Summary: AL-22 Delivery');

    fireEvent.click(contractTab);
    expect(document.querySelector('.artifact-pre')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).toHaveTextContent('Contract: AL-22 Rich Markdown');

    // Switch back to rich mode on contract tab
    fireEvent.click(richBtn);
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'Contract: AL-22 Rich Markdown' })).toBeInTheDocument();

    // Switch to summary tab in rich mode
    fireEvent.click(summaryTab);
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'Summary: AL-22 Delivery' })).toBeInTheDocument();
  });

  it('renders large multiline documents and tables with high node counts and preserves copy fidelity', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    // Generate large document: 100 table rows, 50 task items, multiline code blocks
    const tableHeader = '| Row ID | Metric Name | Threshold | Current Value | Status |\n| :--- | :--- | :--- | :--- | :--- |\n';
    const tableRows = Array.from({ length: 100 }, (_, i) => `| row-${i} | metric_telemetry_${i} | ${i * 10}ms | ${i * 8}ms | healthy |`).join('\n');
    const taskItems = Array.from({ length: 50 }, (_, i) => `- [${i % 2 === 0 ? 'x' : ' '}] Task validation checkpoint ${i}`).join('\n');
    const codeBlock = '```typescript\n' + Array.from({ length: 100 }, (_, i) => `const val_${i} = ${i};`).join('\n') + '\n```';

    const largeDocument = `# Stress Test: Scaled Workload
## Section 1: Massive Telemetry Table
${tableHeader}${tableRows}

## Section 2: Large Task Ledger
${taskItems}

## Section 3: Extended Code Block
${codeBlock}
`;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: largeDocument,
          summary_content: null,
        }}
      />
    );

    // Verify rich render mounts successfully
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'Stress Test: Scaled Workload' })).toBeInTheDocument();

    // Verify all 100 data rows in table are rendered
    const rows = document.querySelectorAll('.markdown-table tbody tr');
    expect(rows.length).toBe(100);

    // Verify all 50 task checkboxes rendered with appropriate checked states
    const checkboxes = screen.getAllByRole('checkbox') as HTMLInputElement[];
    expect(checkboxes.length).toBe(50);
    expect(checkboxes[0]).toBeChecked();
    expect(checkboxes[1]).not.toBeChecked();

    // Verify large document copy fidelity
    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    await act(async () => {
      fireEvent.click(copyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledWith(largeDocument);

    // Verify toggling between views with large content
    const rawBtn = screen.getByRole('button', { name: 'Raw Source' });
    fireEvent.click(rawBtn);
    const pre = document.querySelector('.artifact-pre');
    expect(pre).toBeInTheDocument();
    expect(pre).toHaveTextContent('Stress Test: Scaled Workload');

    const richBtn = screen.getByRole('button', { name: 'Rich Document' });
    fireEvent.click(richBtn);
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(document.querySelectorAll('.markdown-table tbody tr').length).toBe(100);
  });

  it('gracefully handles clipboard write rejection without throwing uncaught errors', async () => {
    const writeTextMock = vi.fn().mockRejectedValue(new Error('Clipboard write permission denied'));
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const copyBtn = screen.getByRole('button', { name: 'Copy' });
    await act(async () => {
      fireEvent.click(copyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(1);
    // Button state remains 'Copy' because error occurred in catch block
    expect(screen.getByRole('button', { name: 'Copy' })).toBeInTheDocument();
  });

  it('correctly renders deeply nested blockquotes and special character sequences', () => {
    const specialContent = `# Special Characters and Nesting
> Level 1 quote
>> Level 2 quote
>>> Level 3 quote
>>>> Level 4 quote

### Entity Characters
Special symbols: & < > " ' / \\ $ ~ % ^ * @ ! #
Math expression: a < b and c > d & e != f

### Strikethrough and Autolink
Here is ~~obsolete text~~ and https://example.com/status
`;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: specialContent,
          summary_content: null,
        }}
      />
    );

    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'Special Characters and Nesting' })).toBeInTheDocument();

    const blockquotes = document.querySelectorAll('.markdown-blockquote');
    expect(blockquotes.length).toBeGreaterThanOrEqual(4);

    expect(screen.getByText(/Special symbols:/)).toBeInTheDocument();
    expect(screen.getByText(/Math expression:/)).toBeInTheDocument();

    // Check autolink
    const link = document.querySelector('a.markdown-link');
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', 'https://example.com/status');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');

    // Check strikethrough (del element)
    const del = document.querySelector('del');
    expect(del).toBeInTheDocument();
    expect(del).toHaveTextContent('obsolete text');
  });

  it('displays empty note in both rich and raw view modes when tab has no content', () => {
    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: '# Non-empty Contract',
          summary_content: null,
        }}
      />
    );

    const rawBtn = screen.getByRole('button', { name: 'Raw Source' });
    const richBtn = screen.getByRole('button', { name: 'Rich Document' });
    const summaryTab = screen.getByRole('button', { name: /summary\.md/i });
    const contractTab = screen.getByRole('button', { name: /contract\.md/i });

    // In rich mode, summary tab shows empty note
    fireEvent.click(summaryTab);
    expect(screen.getByText('No SUMMARY.md document found for this workspace.')).toBeInTheDocument();
    expect(document.querySelector('.artifact-markdown-container')).not.toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).not.toBeInTheDocument();

    // Switch to raw mode while on empty summary tab: still shows empty note
    fireEvent.click(rawBtn);
    expect(screen.getByText('No SUMMARY.md document found for this workspace.')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).not.toBeInTheDocument();

    // Switch back to contract tab: displays raw code block
    fireEvent.click(contractTab);
    expect(document.querySelector('.artifact-pre')).toBeInTheDocument();
    expect(document.querySelector('.artifact-pre')).toHaveTextContent('# Non-empty Contract');

    // Switch back to rich mode: displays rich markdown container
    fireEvent.click(richBtn);
    expect(document.querySelector('.artifact-markdown-container')).toBeInTheDocument();
  });

  it('clears copy timer on unmount and debounces consecutive copy clicks without timer collisions', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    const { unmount } = render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: sampleContract,
          summary_content: sampleSummary,
        }}
      />
    );

    const copyBtn = screen.getByRole('button', { name: 'Copy' });

    // First click: sets Copied
    await act(async () => {
      fireEvent.click(copyBtn);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Advance 1000ms
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Click again at t = 1000ms: resets the 2000ms timer
    await act(async () => {
      fireEvent.click(copyBtn);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Advance 1500ms (t = 2500ms from start, but 1500ms from second click): should still be Copied
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Advance remaining 501ms (t = 2001ms from second click): should revert to Copy
    act(() => {
      vi.advanceTimersByTime(501);
    });
    expect(screen.getByRole('button', { name: 'Copy' })).toBeInTheDocument();

    // Click again and immediately unmount component before timer completes
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Copy' }));
    });
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument();

    // Unmount triggers useEffect cleanup hook without errors
    unmount();
    act(() => {
      vi.advanceTimersByTime(3000);
    });
  });

  it('safely handles malicious inline HTML event handlers and unclosed HTML tags', () => {
    const maliciousHtml = `# Malicious Inline HTML
<button onclick="window.__injected_click__ = true">Click me</button>
<div onmouseover="window.__injected_hover__ = true">Hover me</div>
<span class="unclosed-tag">
<a href="https://example.com" onclick="window.__injected_link_click__ = true">Safe link</a>
`;

    render(
      <ArtifactViewer
        artifacts={{
          ticket_id: 'AL-22',
          contract_content: maliciousHtml,
          summary_content: null,
        }}
      />
    );

    const container = document.querySelector('.artifact-markdown-container');
    expect(container).toBeInTheDocument();

    // Verify raw button with onclick was not injected into the DOM
    const buttons = container?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBe(0);

    // Verify no DOM elements contain onclick or onmouseover attribute handlers
    const allElements = container?.querySelectorAll('*') ?? [];
    allElements.forEach((el) => {
      expect(el.getAttribute('onclick')).toBeNull();
      expect(el.getAttribute('onmouseover')).toBeNull();
    });
  });

  it('renders dynamic configured artifact tabs, displays descriptions, and allows switching', () => {
    const customArtifacts = {
      ticket_id: 'AL-36',
      items: [
        {
          path: 'CONTRACT.md',
          title: 'Technical Contract',
          description: 'Technical constraints',
          content: '# Contract Document\n\nContract rules here.',
        },
        {
          path: 'iterations/01/PLAN.md',
          title: 'Iteration Plan',
          description: 'Step by step plan',
          content: '# Plan Document\n\nPlan tasks here.',
        },
        {
          path: 'iterations/01/REVIEW.md',
          title: 'Code Review',
          description: 'Audit feedback',
          content: null,
        },
      ],
    };

    render(<ArtifactViewer artifacts={customArtifacts} />);

    // Verify all 3 custom tabs rendered with appropriate labels and descriptions
    const contractTab = screen.getByRole('button', { name: 'Technical Contract' });
    expect(contractTab).toBeInTheDocument();
    expect(contractTab).toHaveAttribute('title', 'Technical constraints');

    const planTab = screen.getByRole('button', { name: 'Iteration Plan' });
    expect(planTab).toBeInTheDocument();
    expect(planTab).toHaveAttribute('title', 'Step by step plan');

    const reviewTab = screen.getByRole('button', { name: /Code Review \(Empty\)/i });
    expect(reviewTab).toBeInTheDocument();
    expect(reviewTab).toHaveAttribute('title', 'Audit feedback');

    // First tab content displayed by default
    expect(screen.getByText('Contract rules here.')).toBeInTheDocument();

    // Click second tab
    fireEvent.click(planTab);
    expect(screen.getByText('Plan tasks here.')).toBeInTheDocument();
    expect(screen.queryByText('Contract rules here.')).not.toBeInTheDocument();

    // Click empty tab
    fireEvent.click(reviewTab);
    expect(screen.getByText('No Code Review document found for this workspace.')).toBeInTheDocument();
  });
});
