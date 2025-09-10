package main

import (
	"testing"
)

func TestParseTimeInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        float64
		expectError bool
	}{
		// Original format tests
		{"Seconds as number", "120", 120.0, false},
		{"HH:MM:SS format", "01:30:45", 5445.0, false},
		{"MM:SS format", "05:30", 330.0, false},
		
		// New format tests
		{"Hours with suffix", "2h", 7200.0, false},
		{"Decimal hours", "1.5h", 5400.0, false},
		{"Minutes with suffix", "90m", 5400.0, false},
		{"Decimal minutes", "5.5m", 330.0, false},
		
		// Combined format tests
		{"Hours and minutes", "1h30m", 5400.0, false},
		{"Hours, minutes and seconds", "1h30m15s", 5415.0, false},
		{"Minutes and seconds", "30m15s", 1815.0, false},
		{"Hours and seconds", "2h15s", 7215.0, false},
		
		// TV show length tests
		{"45 minute episode", "45m", 2700.0, false},
		{"60 minute episode", "1h", 3600.0, false},
		{"90 minute episode", "1h30m", 5400.0, false},
		{"2 hour movie", "2h", 7200.0, false},
		
		// Error cases
		{"Invalid format", "invalid", 0.0, true},
		{"Empty string", "", 0.0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeInput(tt.input)
			
			// Check error status
			if (err != nil) != tt.expectError {
				t.Errorf("parseTimeInput(%q) error = %v, expectError %v", 
					tt.input, err, tt.expectError)
				return
			}
			
			// If expecting success, check value
			if !tt.expectError && got != tt.want {
				t.Errorf("parseTimeInput(%q) = %v, want %v", 
					tt.input, got, tt.want)
			}
		})
	}
}
