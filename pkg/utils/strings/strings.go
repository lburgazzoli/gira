package strings

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Truncate shortens a string to the specified maximum length.
// If the string is longer than maxLen, it's truncated and "..." is appended.
// The function is UTF-8 aware and counts runes, not bytes.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	
	if maxLen <= 3 {
		// If maxLen is very small, just return up to maxLen characters without ellipsis
		if utf8.RuneCountInString(s) <= maxLen {
			return s
		}
		return string([]rune(s)[:maxLen])
	}
	
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	
	// Truncate to maxLen-3 to leave room for "..."
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// TruncateBytes is like Truncate but operates on byte length rather than rune count.
// This is useful for cases where byte length matters more than visual character count.
func TruncateBytes(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	
	if maxLen <= 3 {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen]
	}
	
	if len(s) <= maxLen {
		return s
	}
	
	return s[:maxLen-3] + "..."
}

// PrintWrapped writes text to the provided writer, wrapping lines at word boundaries
// to ensure no line exceeds the specified maximum width.
func PrintWrapped(writer io.Writer, text string, maxWidth int) error {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var currentLine strings.Builder
	for _, word := range words {
		// If adding this word would exceed the width, write current line and start new one
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > maxWidth {
			if _, err := writer.Write([]byte(currentLine.String() + "\n")); err != nil {
				return err
			}
			currentLine.Reset()
		}
		
		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}
	
	// Write the last line if it has content
	if currentLine.Len() > 0 {
		if _, err := writer.Write([]byte(currentLine.String() + "\n")); err != nil {
			return err
		}
	}
	
	return nil
}