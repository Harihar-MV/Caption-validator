package parser

import (
	"strings"
	"testing"
)

func TestDetectCaptionFormat(t *testing.T) {
	// These tests are limited since we can't easily create test files
	// In a real implementation, we would use testdata files

	t.Run("Unsupported format returns error", func(t *testing.T) {
		format, err := DetectCaptionFormat("nonexistent.txt")
		if err == nil {
			t.Error("Expected error for unsupported format, got nil")
		}
		if format != "" {
			t.Errorf("Expected empty format string, got %s", format)
		}
	})
}

func TestExtractPlainText(t *testing.T) {
	captions := []Caption{
		{
			Index:     1,
			StartTime: 0,
			EndTime:   5,
			Text:      "This is the first caption.",
		},
		{
			Index:     2,
			StartTime: 5,
			EndTime:   10,
			Text:      "This is the <b>second</b> caption.",
		},
		{
			Index:     3,
			StartTime: 10,
			EndTime:   15,
			Text:      "This is the\nthird caption.",
		},
	}

	expected := "This is the first caption. This is the second caption. This is the\nthird caption."
	result := ExtractPlainText(captions)

	if result != expected {
		t.Errorf("ExtractPlainText failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

func TestParseWebVTT(t *testing.T) {
	webvttContent := `WEBVTT

1
00:00:01.000 --> 00:00:04.000
This is the first caption.

2
00:00:05.000 --> 00:00:09.000
This is the second caption.
`

	reader := strings.NewReader(webvttContent)
	captions, err := parseWebVTT(reader)

	if err != nil {
		t.Fatalf("parseWebVTT returned error: %v", err)
	}

	if len(captions) != 2 {
		t.Fatalf("Expected 2 captions, got %d", len(captions))
	}

	if captions[0].StartTime != 1.0 || captions[0].EndTime != 4.0 {
		t.Errorf("First caption timing incorrect: got %f-->%f", captions[0].StartTime, captions[0].EndTime)
	}

	if captions[0].Text != "This is the first caption." {
		t.Errorf("First caption text incorrect: %s", captions[0].Text)
	}

	if captions[1].StartTime != 5.0 || captions[1].EndTime != 9.0 {
		t.Errorf("Second caption timing incorrect: got %f-->%f", captions[1].StartTime, captions[1].EndTime)
	}
}

func TestParseSRT(t *testing.T) {
	srtContent := `1
00:00:01,000 --> 00:00:04,000
This is the first caption.

2
00:00:05,000 --> 00:00:09,000
This is the second caption.
It has multiple lines.

3
00:00:10,000 --> 00:00:14,000
This is the third caption.
`

	reader := strings.NewReader(srtContent)
	captions, err := parseSRT(reader)

	if err != nil {
		t.Fatalf("parseSRT returned error: %v", err)
	}

	if len(captions) != 3 {
		t.Fatalf("Expected 3 captions, got %d", len(captions))
	}

	if captions[0].StartTime != 1.0 || captions[0].EndTime != 4.0 {
		t.Errorf("First caption timing incorrect: got %f-->%f", captions[0].StartTime, captions[0].EndTime)
	}

	if captions[1].Text != "This is the second caption.\nIt has multiple lines." {
		t.Errorf("Second caption text incorrect: %s", captions[1].Text)
	}

	if captions[2].StartTime != 10.0 || captions[2].EndTime != 14.0 {
		t.Errorf("Third caption timing incorrect: got %f-->%f", captions[2].StartTime, captions[2].EndTime)
	}
}
