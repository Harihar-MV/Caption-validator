package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"caption-validator/internal/client"
	"caption-validator/internal/parser"
	"caption-validator/internal/validator"
)

func main() {
	// Configure logging to file
	logFile := "caption-validator.log"
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	// Parse command line flags
	minCoverage := flag.Float64("coverage", 95.0, "Minimum percentage of time that should be covered by captions")
	tStart := flag.String("t_start", "0", "Start time in seconds or HH:MM:SS format")
	tEnd := flag.String("t_end", "", "End time in seconds or HH:MM:SS format (required)")
	apiURL := flag.String("api", "http://localhost:8080/validate", "URL of the language validation API")
	flag.Parse()

	// Ensure we have a captions file path as the last argument
	args := flag.Args()
	if len(args) != 1 {
		log.Println("Error: Missing captions file path")
		os.Exit(1)
	}
	captionsPath := args[0]

	// Check if file exists
	if _, err := os.Stat(captionsPath); os.IsNotExist(err) {
		log.Printf("Error: Captions file does not exist: %s\n", captionsPath)
		// Print error to stdout in the same format as other errors
		fmt.Printf("{\"type\": \"file_not_found\", \"file\": \"%s\", \"error\": \"Caption file not found\"}\n", captionsPath)
		os.Exit(1)
	}

	// Parse start and end times to seconds
	startSec, err := parseTimeInput(*tStart)
	if err != nil {
		log.Printf("Error parsing t_start: %v\n", err)
		os.Exit(1)
	}

	// End time is required
	if *tEnd == "" {
		log.Println("Error: t_end is required")
		os.Exit(1)
	}

	endSec, err := parseTimeInput(*tEnd)
	if err != nil {
		log.Printf("Error parsing t_end: %v\n", err)
		os.Exit(1)
	}

	// Detect and parse captions file
	captions, format, err := parser.ParseCaptionsFile(captionsPath)
	if err != nil {
		if err == parser.ErrUnsupportedFormat {
			log.Printf("Error: Unsupported caption format for file: %s\n", captionsPath)
			// Print error to stdout to differentiate from success cases
			fmt.Printf("{\"type\": \"unsupported_format\", \"file\": \"%s\", \"error\": \"Unsupported caption file format\"}\n", captionsPath)
			os.Exit(1)
		}
		log.Printf("Error parsing captions file: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Detected caption format: %s\n", format)
	log.Printf("Validating captions from %s to %s with minimum coverage of %.2f%%\n", 
		formatSeconds(startSec), formatSeconds(endSec), *minCoverage)

	// Get the plain text content from captions
	plainText := parser.ExtractPlainText(captions)

	// Perform validations
	hasFailures := false

	// Validate caption coverage
	coverageResult, err := validator.ValidateCoverage(captions, startSec, endSec, *minCoverage)
	if err != nil {
		log.Printf("Error validating coverage: %v\n", err)
		os.Exit(1)
	}
	
	if !coverageResult.Valid {
		fmt.Printf("%s\n", coverageResult.JSON())
		hasFailures = true
	}

	// Validate language via API if URL is provided
	if *apiURL != "" {
		log.Printf("Validating language using API: %s", *apiURL)
		langResult, err := client.ValidateLanguage(*apiURL, plainText)
		if err != nil {
			log.Printf("Error validating language: %v\n", err)
			log.Println("Skipping language validation")
		} else {
			log.Printf("Language validation result: detected='%s', expected='%s', valid=%v", 
				langResult.Language, langResult.ExpectedLang, langResult.Valid)
			if !langResult.Valid {
				fmt.Printf("%s\n", langResult.JSON())
				log.Println("Validation failed: Non-English language detected")
				os.Exit(1)
			}
		}
	} else {
		log.Println("Language validation skipped (no API URL provided)")
	}

	// Exit with code 0 regardless of validation failures
	if hasFailures {
		log.Println("Validation completed with failures")
	} else {
		log.Println("Validation completed successfully")
	}
}

// parseTimeInput converts a time string (either seconds or HH:MM:SS format) to seconds
// Also supports extended formats for longer content like TV shows
func parseTimeInput(timeStr string) (float64, error) {
	// Try to parse as seconds first
	seconds, err := strconv.ParseFloat(timeStr, 64)
	if err == nil {
		return seconds, nil
	}

	// Support for hour notation with h suffix (e.g., "2h", "2.5h")
	if strings.HasSuffix(timeStr, "h") {
		h, err := strconv.ParseFloat(strings.TrimSuffix(timeStr, "h"), 64)
		if err == nil {
			return h * 3600, nil
		}
	}

	// Support for minute notation with m suffix (e.g., "90m")
	if strings.HasSuffix(timeStr, "m") {
		m, err := strconv.ParseFloat(strings.TrimSuffix(timeStr, "m"), 64)
		if err == nil {
			return m * 60, nil
		}
	}

	// Support for extended time format (e.g., "2h30m15s")
	if strings.Contains(timeStr, "h") || strings.Contains(timeStr, "m") || strings.Contains(timeStr, "s") {
		total := 0.0
		
		// Extract hours
		hParts := strings.Split(timeStr, "h")
		if len(hParts) > 1 {
			h, err := strconv.ParseFloat(hParts[0], 64)
			if err == nil {
				total += h * 3600
			}
			timeStr = hParts[1]
		}
		
		// Extract minutes
		mParts := strings.Split(timeStr, "m")
		if len(mParts) > 1 {
			m, err := strconv.ParseFloat(mParts[0], 64)
			if err == nil {
				total += m * 60
			}
			timeStr = mParts[1]
		}
		
		// Extract seconds
		sParts := strings.Split(timeStr, "s")
		if len(sParts) > 0 {
			s, err := strconv.ParseFloat(sParts[0], 64)
			if err == nil {
				total += s
			}
		}
		
		if total > 0 {
			return total, nil
		}
	}

	// Try to parse as HH:MM:SS or MM:SS
	parts := strings.Split(timeStr, ":")
	if len(parts) == 3 {
		// HH:MM:SS
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %v", err)
		}
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %v", err)
		}
		s, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %v", err)
		}
		return float64(h*3600 + m*60) + s, nil
	} else if len(parts) == 2 {
		// MM:SS
		m, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %v", err)
		}
		s, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %v", err)
		}
		return float64(m*60) + s, nil
	}

	return 0, fmt.Errorf("invalid time format: %s", timeStr)
}

// formatSeconds converts seconds to HH:MM:SS format
func formatSeconds(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := seconds - float64(h*3600+m*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", h, m, s)
}
