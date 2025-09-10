package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// parseWebVTT parses a WebVTT format file
func parseWebVTT(r io.Reader) ([]Caption, error) {
	var captions []Caption
	scanner := bufio.NewScanner(r)

	// First line should be "WEBVTT"
	if !scanner.Scan() {
		return nil, errors.New("empty file")
	}
	
	if !strings.HasPrefix(scanner.Text(), "WEBVTT") {
		return nil, errors.New("missing WEBVTT header")
	}

	// Skip header section until we find an empty line
	for scanner.Scan() {
		if scanner.Text() == "" {
			break
		}
	}

	var currentCaption Caption
	inCaption := false
	textLines := []string{}
	index := 1

	// Parse cues
	for scanner.Scan() {
		line := scanner.Text()

		// Empty line indicates the end of a caption
		if line == "" {
			if inCaption {
				currentCaption.Text = strings.Join(textLines, "\n")
				currentCaption.Index = index
				captions = append(captions, currentCaption)
				textLines = []string{}
				index++
				inCaption = false
			}
			continue
		}

		// Check for time line (containing "-->")
		if strings.Contains(line, "-->") {
			if inCaption {
				// This shouldn't happen, but let's handle it gracefully
				currentCaption.Text = strings.Join(textLines, "\n")
				currentCaption.Index = index
				captions = append(captions, currentCaption)
				textLines = []string{}
				index++
			}

			// Parse time codes
			startTime, endTime, err := parseWebVTTTimeline(line)
			if err != nil {
				return nil, err
			}

			currentCaption = Caption{
				StartTime: startTime,
				EndTime:   endTime,
			}
			inCaption = true
		} else if inCaption {
			// This is caption text
			textLines = append(textLines, line)
		}
		// Ignore any lines that aren't part of a caption (could be comments, etc.)
	}

	// Handle the last caption if we were in one
	if inCaption {
		currentCaption.Text = strings.Join(textLines, "\n")
		currentCaption.Index = index
		captions = append(captions, currentCaption)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return captions, nil
}

// parseWebVTTTimeline parses a WebVTT timestamp line
// Example: "00:00:10.500 --> 00:00:13.000"
func parseWebVTTTimeline(line string) (float64, float64, error) {
	// Clean up any leading/trailing whitespace
	line = strings.TrimSpace(line)
	
	// Split on the arrow
	parts := strings.Split(line, "-->")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid time format")
	}
	
	// Parse timestamps
	startTimeStr := strings.TrimSpace(parts[0])
	endTimeStr := strings.TrimSpace(parts[1])
	
	// Extract timestamp from settings if present
	endTimeStr = strings.Split(endTimeStr, " ")[0]
	
	startTime, err := parseWebVTTTimestamp(startTimeStr)
	if err != nil {
		return 0, 0, err
	}
	
	endTime, err := parseWebVTTTimestamp(endTimeStr)
	if err != nil {
		return 0, 0, err
	}
	
	return startTime, endTime, nil
}

// parseWebVTTTimestamp converts a WebVTT timestamp to seconds
// Format: "HH:MM:SS.mmm" or "MM:SS.mmm"
func parseWebVTTTimestamp(timestamp string) (float64, error) {
	parts := strings.Split(timestamp, ":")
	
	var hours, minutes, seconds float64
	var err error
	
	if len(parts) == 3 {
		// HH:MM:SS.mmm
		hours, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		
		minutes, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		
		seconds, err = strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
		
		return hours*3600 + minutes*60 + seconds, nil
	} else if len(parts) == 2 {
		// MM:SS.mmm
		minutes, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		
		seconds, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		
		return minutes*60 + seconds, nil
	}
	
	return 0, errors.New("invalid timestamp format")
}
