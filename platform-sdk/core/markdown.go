package core

import (
	"regexp"
	"strings"
)

var (
	reCodeBlock   = regexp.MustCompile("(?s)```[a-zA-Z]*\n?(.*?)```")
	reInlineCode  = regexp.MustCompile("`([^`]+)`")
	reBoldAst     = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reBoldUnd     = regexp.MustCompile(`__(.+?)__`)
	reItalicAst   = regexp.MustCompile(`\*(.+?)\*`)
	reItalicUnd   = regexp.MustCompile(`_(.+?)_`)
	reStrike      = regexp.MustCompile(`~~(.+?)~~`)
	reLink        = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reHeading     = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	reHorizontal  = regexp.MustCompile(`(?m)^---+\s*$`)
	reBlockquote  = regexp.MustCompile(`(?m)^>\s?`)
)

// StripMarkdown converts Markdown-formatted text to clean plain text.
// Useful for platforms that don't support Markdown rendering.
func StripMarkdown(s string) string {
	// Preserve code block content but remove fences
	s = reCodeBlock.ReplaceAllString(s, "$1")

	// Inline code — remove backticks
	s = reInlineCode.ReplaceAllString(s, "$1")

	// Bold / italic / strikethrough — keep text
	s = reBoldAst.ReplaceAllString(s, "$1")
	s = reBoldUnd.ReplaceAllString(s, "$1")
	s = reItalicAst.ReplaceAllString(s, "$1")
	s = reItalicUnd.ReplaceAllString(s, "$1")
	s = reStrike.ReplaceAllString(s, "$1")

	// Links [text](url) → text (url)
	s = reLink.ReplaceAllString(s, "$1 ($2)")

	// Headings — remove # prefix
	s = reHeading.ReplaceAllString(s, "")

	// Horizontal rules
	s = reHorizontal.ReplaceAllString(s, "")

	// Blockquotes
	s = reBlockquote.ReplaceAllString(s, "")

	// Collapse 3+ consecutive blank lines into 2
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}

// SplitMessageCodeFenceAware splits text into chunks of maxLen,
// being careful not to split in the middle of code blocks.
func SplitMessageCodeFenceAware(s string, maxLen int) []string {
	if len(s) <= maxLen {
		return []string{s}
	}

	var parts []string
	remaining := s

	for len(remaining) > 0 {
		if len(remaining) <= maxLen {
			parts = append(parts, remaining)
			break
		}

		// Try to split at a natural boundary
		splitAt := maxLen

		// Don't split in the middle of a code block
		codeStart := strings.LastIndex(remaining[:splitAt], "```")
		if codeStart >= 0 {
			// Check if we're inside a code block
			codeEnd := strings.Index(remaining[codeStart+3:], "```")
			if codeEnd >= 0 && codeStart+3+codeEnd > splitAt {
				// We're inside a code block, find a safer place to split
				if codeStart > 0 {
					splitAt = codeStart
				}
			}
		}

		// Prefer to split at a double newline
		if idx := strings.LastIndex(remaining[:splitAt], "\n\n"); idx > 0 {
			splitAt = idx + 2
		} else if idx := strings.LastIndex(remaining[:splitAt], "\n"); idx > 0 {
			// Or at a single newline
			splitAt = idx + 1
		}

		parts = append(parts, remaining[:splitAt])
		remaining = remaining[splitAt:]
	}

	return parts
}

// MarkdownToSimpleHTML converts a limited subset of Markdown to HTML for platforms
// that support basic HTML formatting (Telegram, Discord, etc.).
func MarkdownToSimpleHTML(s string) string {
	// Order matters: code blocks first to protect their content
	s = escapeHTML(s)

	// Code blocks ```code``` → <pre>code</pre>
	s = regexp.MustCompile("(?s)```([a-zA-Z]*)\n?(.*?)```").ReplaceAllString(s, "<pre>$2</pre>")

	// Inline code `code` → <code>code</code>
	s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "<code>$1</code>")

	// Bold **text** or __text__ → <b>text</b>
	s = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(s, "<b>$1</b>")
	s = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(s, "<b>$1</b>")

	// Italic *text* or _text_ → <i>text</i>
	s = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(s, "<i>$1</i>")
	s = regexp.MustCompile(`_(.+?)_`).ReplaceAllString(s, "<i>$1</i>")

	// Strikethrough ~~text~~ → <s>text</s>
	s = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(s, "<s>$1</s>")

	// Links [text](url) → <a href="url">text</a>
	s = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(s, `<a href="$2">$1</a>`)

	// Headings # Heading → <b>Heading</b>
	s = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`).ReplaceAllString(s, "<b>$1</b>")

	// Blockquotes > text → <i>&gt; text</i>
	s = regexp.MustCompile(`(?m)^>\s?(.+)$`).ReplaceAllString(s, "<i>&gt; $1</i>")

	return s
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
)

func escapeHTML(s string) string {
	return htmlReplacer.Replace(s)
}
