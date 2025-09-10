package parser

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Supported caption formats
const (
	FormatWebVTT = "WebVTT"
	FormatSRT    = "SRT"
)

// Errors
var (
	ErrUnsupportedFormat = errors.New("unsupported caption format")
)

// Caption represents a single caption entry
type Caption struct {
	Index     int
	StartTime float64 // in seconds
	EndTime   float64 // in seconds
	Text      string
}

// DetectCaptionFormat determines the format of a captions file
func DetectCaptionFormat(filePath string) (string, error) {
	// Read the first few bytes to check magic bytes or signature
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check file extension as a hint
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Read first 100 bytes for format detection
	header := make([]byte, 100)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	// Try using the 'file' command as a backup
	cmd := exec.Command("file", filePath)
	output, cmdErr := cmd.Output()
	fileType := ""
	if cmdErr == nil {
		fileType = strings.ToLower(string(output))
	}

	// Check for WebVTT signature
	if bytes.HasPrefix(header, []byte("WEBVTT")) || strings.Contains(string(header), "WEBVTT") {
		return FormatWebVTT, nil
	}
	
	// Check for SRT format
	// SRT files typically start with a number (index), followed by time codes with arrow
	if ext == ".srt" || strings.Contains(fileType, "subrip") || 
		regexp.MustCompile(`^\d+\s*\r?\n\d{2}:\d{2}:\d{2},\d{3}\s*-->`).Match(header) {
		return FormatSRT, nil
	}

	return "", ErrUnsupportedFormat
}

// ParseCaptionsFile detects and parses a captions file
func ParseCaptionsFile(filePath string) ([]Caption, string, error) {
	format, err := DetectCaptionFormat(filePath)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	var captions []Caption

	switch format {
	case FormatWebVTT:
		captions, err = parseWebVTT(file)
	case FormatSRT:
		captions, err = parseSRT(file)
	default:
		return nil, "", ErrUnsupportedFormat
	}

	if err != nil {
		return nil, "", err
	}

	return captions, format, nil
}

// ExtractPlainText gets all text content from captions
func ExtractPlainText(captions []Caption) string {
	var builder strings.Builder
	
	for _, caption := range captions {
		// Remove HTML tags if present
		text := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(caption.Text, "")
		builder.WriteString(text)
		builder.WriteString(" ")
	}
	
	return strings.TrimSpace(builder.String())
}
