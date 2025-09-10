package parser

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

// ParseLargeCaptionsFile is optimized for large caption files by using chunked processing
func ParseLargeCaptionsFile(filePath string) ([]Caption, string, error) {
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
		captions, err = parseChunkedWebVTT(file)
	case FormatSRT:
		captions, err = parseChunkedSRT(file)
	default:
		return nil, "", ErrUnsupportedFormat
	}

	if err != nil {
		return nil, "", err
	}

	return captions, format, nil
}

// parseChunkedWebVTT parses WebVTT files in chunks to reduce memory usage
func parseChunkedWebVTT(r io.Reader) ([]Caption, error) {
	var captions []Caption
	scanner := bufio.NewScanner(r)
	bufSize := 64 * 1024 // 64KB buffer
	scanner.Buffer(make([]byte, bufSize), bufSize)

	// First line should be "WEBVTT"
	if !scanner.Scan() {
		return nil, io.EOF
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

	// Process in chunks of 100 captions to avoid excessive memory usage
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

// parseChunkedSRT parses SRT files in chunks to reduce memory usage
func parseChunkedSRT(r io.Reader) ([]Caption, error) {
	var captions []Caption
	scanner := bufio.NewScanner(r)
	bufSize := 64 * 1024 // 64KB buffer
	scanner.Buffer(make([]byte, bufSize), bufSize)
	
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
				if strings.Contains(trimmedLine, "-->") {
					startTime, endTime, timeErr := parseSRTTimeline(trimmedLine)
					if timeErr == nil {
						currentCaption = Caption{Index: len(captions) + 1, StartTime: startTime, EndTime: endTime}
						parseState = 2 // Skip to text parsing
						continue
					}
				}
				// Otherwise, this could be text from a previous caption with a missing blank line
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
