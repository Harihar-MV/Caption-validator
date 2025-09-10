package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestParseLargeCaptionsFile(t *testing.T) {
	// Test with sample WebVTT content
	webvttContent := `WEBVTT

1
00:00:01.000 --> 00:00:05.000
This is the first caption

2
00:00:06.000 --> 00:00:10.000
This is the second caption
spanning multiple lines

3
00:01:00.000 --> 00:01:30.000
This is a longer caption
`

	// Test with sample SRT content
	srtContent := `1
00:00:01,000 --> 00:00:05,000
This is the first caption

2
00:00:06,000 --> 00:00:10,000
This is the second caption
spanning multiple lines

3
00:01:00,000 --> 00:01:30,000
This is a longer caption
`

	// Create temporary test files
	tmpDir := t.TempDir()
	webvttFile := filepath.Join(tmpDir, "test.vtt")
	srtFile := filepath.Join(tmpDir, "test.srt")

	// Write test data to files
	if err := os.WriteFile(webvttFile, []byte(webvttContent), 0644); err != nil {
		t.Fatalf("Failed to create test WebVTT file: %v", err)
	}
	if err := os.WriteFile(srtFile, []byte(srtContent), 0644); err != nil {
		t.Fatalf("Failed to create test SRT file: %v", err)
	}

	// Test WebVTT parsing
	t.Run("ParseLargeCaptionsFile WebVTT", func(t *testing.T) {
		captions, format, err := ParseLargeCaptionsFile(webvttFile)
		if err != nil {
			t.Fatalf("ParseLargeCaptionsFile returned error: %v", err)
		}
		
		if format != FormatWebVTT {
			t.Errorf("Expected format %s, got %s", FormatWebVTT, format)
		}
		
		if len(captions) != 3 {
			t.Errorf("Expected 3 captions, got %d", len(captions))
		}
		
		// Check first caption
		if captions[0].StartTime != 1.0 || captions[0].EndTime != 5.0 {
			t.Errorf("First caption time mismatch: got %f-%f, want 1.0-5.0", 
				captions[0].StartTime, captions[0].EndTime)
		}
		
		// Check last caption
		if captions[2].StartTime != 60.0 || captions[2].EndTime != 90.0 {
			t.Errorf("Last caption time mismatch: got %f-%f, want 60.0-90.0", 
				captions[2].StartTime, captions[2].EndTime)
		}
	})
	
	// Test SRT parsing
	t.Run("ParseLargeCaptionsFile SRT", func(t *testing.T) {
		captions, format, err := ParseLargeCaptionsFile(srtFile)
		if err != nil {
			t.Fatalf("ParseLargeCaptionsFile returned error: %v", err)
		}
		
		if format != FormatSRT {
			t.Errorf("Expected format %s, got %s", FormatSRT, format)
		}
		
		if len(captions) != 3 {
			t.Errorf("Expected 3 captions, got %d", len(captions))
		}
		
		// Check first caption
		if captions[0].StartTime != 1.0 || captions[0].EndTime != 5.0 {
			t.Errorf("First caption time mismatch: got %f-%f, want 1.0-5.0", 
				captions[0].StartTime, captions[0].EndTime)
		}
		
		// Check last caption
		if captions[2].StartTime != 60.0 || captions[2].EndTime != 90.0 {
			t.Errorf("Last caption time mismatch: got %f-%f, want 60.0-90.0", 
				captions[2].StartTime, captions[2].EndTime)
		}
	})
	
	// Test with unsupported format
	t.Run("ParseLargeCaptionsFile Unsupported Format", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.txt")
		if err := os.WriteFile(invalidFile, []byte("This is not a caption file"), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}
		
		_, _, err := ParseLargeCaptionsFile(invalidFile)
		if err == nil {
			t.Error("Expected error for unsupported format, got nil")
		}
	})
	
	// Test with non-existent file
	t.Run("ParseLargeCaptionsFile Non-existent File", func(t *testing.T) {
		_, _, err := ParseLargeCaptionsFile(filepath.Join(tmpDir, "nonexistent.vtt"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

// Test chunked WebVTT parsing directly
func TestParseChunkedWebVTT(t *testing.T) {
	content := `WEBVTT

1
00:00:01.000 --> 00:00:05.000
This is the first caption

2
00:00:06.000 --> 00:00:10.000
This is the second caption
spanning multiple lines
`
	reader := bytes.NewBufferString(content)
	captions, err := parseChunkedWebVTT(reader)
	
	if err != nil {
		t.Fatalf("parseChunkedWebVTT returned error: %v", err)
	}
	
	if len(captions) != 2 {
		t.Errorf("Expected 2 captions, got %d", len(captions))
	}
	
	if captions[0].Text != "This is the first caption" {
		t.Errorf("Caption text mismatch: got %q", captions[0].Text)
	}
	
	if captions[1].Text != "This is the second caption\nspanning multiple lines" {
		t.Errorf("Caption text mismatch: got %q", captions[1].Text)
	}
}

// Test chunked SRT parsing directly
func TestParseChunkedSRT(t *testing.T) {
	content := `1
00:00:01,000 --> 00:00:05,000
This is the first caption

2
00:00:06,000 --> 00:00:10,000
This is the second caption
spanning multiple lines
`
	reader := bytes.NewBufferString(content)
	captions, err := parseChunkedSRT(reader)
	
	if err != nil {
		t.Fatalf("parseChunkedSRT returned error: %v", err)
	}
	
	if len(captions) != 2 {
		t.Errorf("Expected 2 captions, got %d", len(captions))
	}
	
	if captions[0].Text != "This is the first caption" {
		t.Errorf("Caption text mismatch: got %q", captions[0].Text)
	}
	
	if captions[1].Text != "This is the second caption\nspanning multiple lines" {
		t.Errorf("Caption text mismatch: got %q", captions[1].Text)
	}
}
