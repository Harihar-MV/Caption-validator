package parser

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// parseSRT parses a SubRip Text (SRT) format file
func parseSRT(r io.Reader) ([]Caption, error) {
	var captions []Caption
	scanner := bufio.NewScanner(r)
	
	var currentCaption Caption
	var textLines []string
	parseState := 0 // 0=index, 1=timestamp, 2=text

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		
		// Empty line means end of a caption block (unless we're at the beginning)
		if trimmedLine == "" {
			if parseState > 0 {
				// Finalize the current caption if we have one
				if len(textLines) > 0 {
					currentCaption.Text = strings.Join(textLines, "\n")
					captions = append(captions, currentCaption)
					textLines = nil
				}
				parseState = 0
			}
			continue
		}
		
		switch parseState {
		case 0: // Expecting index number
			index, err := strconv.Atoi(trimmedLine)
			if err != nil {
				// If not a number, this might be a timestamp in a malformed file
				// Try to handle it as a timestamp
				if strings.Contains(trimmedLine, "-->") {
					startTime, endTime, timeErr := parseSRTTimeline(trimmedLine)
					if timeErr == nil {
						currentCaption = Caption{Index: len(captions) + 1, StartTime: startTime, EndTime: endTime}
						parseState = 2 // Skip to text parsing
						continue
					}
				}
				// Otherwise, this could be text from a previous caption with a missing blank line
				// Just treat it as text from the previous caption if possible
				if len(captions) > 0 {
					parseState = 2
					textLines = append(textLines, line)
					continue
				}
			} else {
				currentCaption = Caption{Index: index}
				parseState = 1
			}
			
		case 1: // Expecting timestamp line
			if strings.Contains(trimmedLine, "-->") {
				startTime, endTime, err := parseSRTTimeline(trimmedLine)
				if err != nil {
					return nil, err
				}
				currentCaption.StartTime = startTime
				currentCaption.EndTime = endTime
				parseState = 2
				textLines = nil
			} else {
				// Unexpected format, try to recover
				textLines = append(textLines, line)
				parseState = 2
			}
			
		case 2: // Caption text
			textLines = append(textLines, line)
		}
	}
	
	// Handle the last caption if we were parsing one
	if parseState > 0 && len(textLines) > 0 {
		currentCaption.Text = strings.Join(textLines, "\n")
		captions = append(captions, currentCaption)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return captions, nil
}

// parseSRTTimeline parses an SRT timestamp line
// Example: "00:00:10,500 --> 00:00:13,000"
func parseSRTTimeline(line string) (float64, float64, error) {
	// Clean up any leading/trailing whitespace
	line = strings.TrimSpace(line)
	
	// Split on the arrow
	parts := strings.Split(line, "-->")
	if len(parts) != 2 {
		return 0, 0, nil
	}
	
	// Parse timestamps
	startTimeStr := strings.TrimSpace(parts[0])
	endTimeStr := strings.TrimSpace(parts[1])
	
	startTime, err := parseSRTTimestamp(startTimeStr)
	if err != nil {
		return 0, 0, err
	}
	
	endTime, err := parseSRTTimestamp(endTimeStr)
	if err != nil {
		return 0, 0, err
	}
	
	return startTime, endTime, nil
}

// parseSRTTimestamp converts an SRT timestamp to seconds
// Format: "HH:MM:SS,mmm"
func parseSRTTimestamp(timestamp string) (float64, error) {
	// Replace comma with period for milliseconds
	timestamp = strings.Replace(timestamp, ",", ".", 1)
	
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		return 0, nil
	}
	
	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	
	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}
	
	return hours*3600 + minutes*60 + seconds, nil
}
