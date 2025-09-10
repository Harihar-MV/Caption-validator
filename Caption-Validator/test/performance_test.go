package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"caption-validator/internal/parser"
)

// generateLargeCaptionFile creates a test caption file with the specified number of captions
func generateLargeCaptionFile(t *testing.T, numCaptions int, format string) string {
	t.Helper()
	
	var extension, header, captionTemplate string
	
	if format == parser.FormatWebVTT {
		extension = ".vtt"
		header = "WEBVTT\n\n"
		captionTemplate = "%d\n%02d:%02d:%02d.000 --> %02d:%02d:%02d.000\nThis is caption number %d\n\n"
	} else {
		extension = ".srt"
		header = ""
		captionTemplate = "%d\n%02d:%02d:%02d,000 --> %02d:%02d:%02d,000\nThis is caption number %d\n\n"
	}
	
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, fmt.Sprintf("large_captions_%d%s", numCaptions, extension))
	
	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer f.Close()
	
	// Write header if needed
	if header != "" {
		if _, err := f.WriteString(header); err != nil {
			t.Fatalf("Failed to write header: %v", err)
		}
	}
	
	// Generate captions with sequential timestamps
	for i := 1; i <= numCaptions; i++ {
		startSeconds := (i - 1) * 5
		endSeconds := i * 5
		
		startHour := startSeconds / 3600
		startMin := (startSeconds % 3600) / 60
		startSec := startSeconds % 60
		
		endHour := endSeconds / 3600
		endMin := (endSeconds % 3600) / 60
		endSec := endSeconds % 60
		
		caption := fmt.Sprintf(captionTemplate,
			i, startHour, startMin, startSec, endHour, endMin, endSec, i)
		
		if _, err := f.WriteString(caption); err != nil {
			t.Fatalf("Failed to write caption: %v", err)
		}
	}
	
	return filename
}

func BenchmarkParseStandard(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("WebVTT_%d", size), func(b *testing.B) {
			filename := generateLargeCaptionFile(b, size, parser.FormatWebVTT)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, _, err := parser.ParseCaptionsFile(filename)
				if err != nil {
					b.Fatalf("Failed to parse: %v", err)
				}
			}
		})
		
		b.Run(fmt.Sprintf("SRT_%d", size), func(b *testing.B) {
			filename := generateLargeCaptionFile(b, size, parser.FormatSRT)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, _, err := parser.ParseCaptionsFile(filename)
				if err != nil {
					b.Fatalf("Failed to parse: %v", err)
				}
			}
		})
	}
}

func BenchmarkParseLarge(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("WebVTT_%d", size), func(b *testing.B) {
			filename := generateLargeCaptionFile(b, size, parser.FormatWebVTT)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, _, err := parser.ParseLargeCaptionsFile(filename)
				if err != nil {
					b.Fatalf("Failed to parse: %v", err)
				}
			}
		})
		
		b.Run(fmt.Sprintf("SRT_%d", size), func(b *testing.B) {
			filename := generateLargeCaptionFile(b, size, parser.FormatSRT)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, _, err := parser.ParseLargeCaptionsFile(filename)
				if err != nil {
					b.Fatalf("Failed to parse: %v", err)
				}
			}
		})
	}
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}
	
	sizes := []int{1000, 10000, 50000}
	
	for _, size := range sizes {
		t.Run(fmt.Sprintf("Compare_%d_Captions", size), func(t *testing.T) {
			filename := generateLargeCaptionFile(t, size, parser.FormatWebVTT)
			
			// Test standard parsing
			start := time.Now()
			_, _, err := parser.ParseCaptionsFile(filename)
			if err != nil {
				t.Fatalf("Failed to parse with standard parser: %v", err)
			}
			standardTime := time.Since(start)
			
			// Test large file parsing
			start = time.Now()
			_, _, err = parser.ParseLargeCaptionsFile(filename)
			if err != nil {
				t.Fatalf("Failed to parse with large parser: %v", err)
			}
			largeTime := time.Since(start)
			
			t.Logf("Size: %d captions, Standard: %v, Large: %v, Ratio: %.2f", 
				size, standardTime, largeTime, float64(standardTime)/float64(largeTime))
		})
	}
}
