package validator

import (
	"encoding/json"
	"fmt"
	"math"

	"caption-validator/internal/parser"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Valid bool
	Type  string
	Data  map[string]interface{}
}

// JSON returns the JSON representation of the validation result
func (vr ValidationResult) JSON() string {
	result := map[string]interface{}{
		"type": vr.Type,
	}
	
	// Add all other fields from Data
	for k, v := range vr.Data {
		result[k] = v
	}
	
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf(`{"type": "%s", "error": "Error marshalling JSON"}`, vr.Type)
	}
	
	return string(jsonBytes)
}

// ValidateCoverage checks if captions cover the required percentage of time
func ValidateCoverage(captions []parser.Caption, startTime float64, endTime float64, minCoverage float64) (ValidationResult, error) {
	if endTime <= startTime {
		return ValidationResult{}, fmt.Errorf("end time must be greater than start time")
	}
	
	// Calculate total time range
	totalTime := endTime - startTime
	
	// Track covered time segments
	type timeSegment struct {
		start float64
		end   float64
	}
	var coveredSegments []timeSegment
	
	// Add all caption time segments that overlap with our range
	for _, caption := range captions {
		// Skip captions outside our range
		if caption.EndTime <= startTime || caption.StartTime >= endTime {
			continue
		}
		
		// Clip caption to our range
		segStart := math.Max(caption.StartTime, startTime)
		segEnd := math.Min(caption.EndTime, endTime)
		
		coveredSegments = append(coveredSegments, timeSegment{segStart, segEnd})
	}
	
	// Merge overlapping segments
	if len(coveredSegments) > 0 {
		// Sort segments by start time
		for i := 0; i < len(coveredSegments)-1; i++ {
			for j := i + 1; j < len(coveredSegments); j++ {
				if coveredSegments[i].start > coveredSegments[j].start {
					coveredSegments[i], coveredSegments[j] = coveredSegments[j], coveredSegments[i]
				}
			}
		}
		
		// Merge overlapping segments
		merged := []timeSegment{coveredSegments[0]}
		for i := 1; i < len(coveredSegments); i++ {
			last := &merged[len(merged)-1]
			current := coveredSegments[i]
			
			// If current segment overlaps with last merged segment, extend last segment
			if current.start <= last.end {
				if current.end > last.end {
					last.end = current.end
				}
			} else {
				// No overlap, add as new segment
				merged = append(merged, current)
			}
		}
		
		coveredSegments = merged
	}
	
	// Calculate total covered time
	coveredTime := 0.0
	for _, seg := range coveredSegments {
		coveredTime += seg.end - seg.start
	}
	
	// Calculate coverage percentage
	coveragePercent := (coveredTime / totalTime) * 100.0
	
	// Create validation result
	valid := coveragePercent >= minCoverage
	
	result := ValidationResult{
		Valid: valid,
		Type:  "caption_coverage",
		Data: map[string]interface{}{
			"required_coverage": minCoverage,
			"actual_coverage":   math.Round(coveragePercent*100) / 100, // Round to 2 decimal places
			"start_time":        startTime,
			"end_time":          endTime,
			"covered_time":      math.Round(coveredTime*100) / 100,
			"total_time":        totalTime,
		},
	}
	
	if !valid {
		result.Data["missing_coverage_seconds"] = math.Round(((minCoverage/100)*totalTime-coveredTime)*100) / 100
	}
	
	return result, nil
}
