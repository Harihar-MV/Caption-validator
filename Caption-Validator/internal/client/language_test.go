package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateLanguage(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		responseJSON  string
		responseCode  int
		expectValid   bool
		expectLang    string
		expectError   bool
	}{
		{
			name:          "Valid language response",
			responseJSON:  `{"lang": "en-US"}`,
			responseCode:  http.StatusOK,
			expectValid:   true,
			expectLang:    "en-US",
			expectError:   false,
		},
		{
			name:          "Invalid language response",
			responseJSON:  `{"lang": "es-ES"}`,
			responseCode:  http.StatusOK,
			expectValid:   false,
			expectLang:    "es-ES",
			expectError:   false,
		},
		{
			name:          "Server error",
			responseJSON:  `{"error": "Internal server error"}`,
			responseCode:  http.StatusInternalServerError,
			expectValid:   false,
			expectLang:    "",
			expectError:   true,
		},
		{
			name:          "Malformed response",
			responseJSON:  `not valid json`,
			responseCode:  http.StatusOK,
			expectValid:   false,
			expectLang:    "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and content type
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "text/plain" {
					t.Errorf("Expected Content-Type: text/plain, got %s", r.Header.Get("Content-Type"))
				}
				
				// Set response code and write response
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseJSON))
			}))
			defer server.Close()

			// Call the function with the mock server URL
			result, err := ValidateLanguage(server.URL, "Sample caption text")

			// Check for errors
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateLanguage() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Skip further checks if we expected an error
			if tt.expectError {
				return
			}

			// Check validation result
			if result.Valid != tt.expectValid {
				t.Errorf("ValidateLanguage() valid = %v, want %v", result.Valid, tt.expectValid)
			}

			if result.Language != tt.expectLang {
				t.Errorf("ValidateLanguage() language = %v, want %v", result.Language, tt.expectLang)
			}

			// Test JSON output
			jsonStr := result.JSON()
			var jsonObj map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
				t.Fatalf("Failed to parse JSON result: %v", err)
			}

			if jsonObj["type"] != "incorrect_language" {
				t.Errorf("JSON output has incorrect type: %v", jsonObj["type"])
			}

			if jsonObj["detected"] != result.Language {
				t.Errorf("JSON output has incorrect detected language: %v", jsonObj["detected"])
			}

			if jsonObj["expected"] != "en-US" {
				t.Errorf("JSON output has incorrect expected language: %v", jsonObj["expected"])
			}
		})
	}
}
