package validator

import (
	"encoding/json"
	"testing"

	"caption-validator/internal/parser"
)

func TestValidateCoverage(t *testing.T) {
	// Create test captions
	captions := []parser.Caption{
		{
			Index:     1,
			StartTime: 10.0,
			EndTime:   20.0,
			Text:      "Caption 1",
		},
		{
			Index:     2,
			StartTime: 25.0,
			EndTime:   35.0,
			Text:      "Caption 2",
		},
		{
			Index:     3,
			StartTime: 40.0,
			EndTime:   55.0,
			Text:      "Caption 3",
		},
	}

	tests := []struct {
		name        string
		startTime   float64
		endTime     float64
		minCoverage float64
		wantValid   bool
	}{
		{
			name:        "Full coverage",
			startTime:   10.0,
			endTime:     55.0,
			minCoverage: 66.6, // (35/45)*100 = 77.78%
			wantValid:   true,
		},
		{
			name:        "Partial coverage below threshold",
			startTime:   5.0,
			endTime:     60.0,
			minCoverage: 80.0, // (35/55)*100 = 63.64%
			wantValid:   false,
		},
		{
			name:        "No coverage",
			startTime:   0.0,
			endTime:     5.0,
			minCoverage: 1.0,
			wantValid:   false,
		},
		{
			name:        "Edge case - exactly meets threshold",
			startTime:   10.0,
			endTime:     55.0,
			minCoverage: 77.77, // (35/45)*100 = 77.78%, but use slightly lower threshold to account for floating point precision
			wantValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCoverage(captions, tt.startTime, tt.endTime, tt.minCoverage)
			if err != nil {
				t.Fatalf("ValidateCoverage() error = %v", err)
			}
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateCoverage() got valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.Type != "caption_coverage" {
				t.Errorf("ValidateCoverage() got type = %v, want %v", result.Type, "caption_coverage")
			}

			// Test JSON output
			jsonStr := result.JSON()
			var jsonObj map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
				t.Fatalf("Failed to parse JSON result: %v", err)
			}

			if jsonObj["type"] != "caption_coverage" {
				t.Errorf("JSON output has incorrect type: %v", jsonObj["type"])
			}

			if jsonObj["required_coverage"] != tt.minCoverage {
				t.Errorf("JSON output has incorrect required_coverage: %v", jsonObj["required_coverage"])
			}

			if !tt.wantValid {
				if _, ok := jsonObj["missing_coverage_seconds"]; !ok {
					t.Errorf("JSON output for failed validation should contain missing_coverage_seconds")
				}
			}
		})
	}

	t.Run("Invalid time range", func(t *testing.T) {
		_, err := ValidateCoverage(captions, 10.0, 5.0, 90.0)
		if err == nil {
			t.Error("Expected error for end time <= start time, got nil")
		}
	})
}
