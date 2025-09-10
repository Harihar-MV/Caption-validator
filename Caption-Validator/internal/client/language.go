package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LanguageResponse represents the response from language validation API
type LanguageResponse struct {
	Lang string `json:"lang"`
}

// LanguageValidationResult represents the result of language validation
type LanguageValidationResult struct {
	Valid        bool
	Type         string
	Language     string
	ExpectedLang string
}

// JSON returns the JSON representation of the validation result
func (lvr LanguageValidationResult) JSON() string {
	result := map[string]interface{}{
		"type":            "incorrect_language",
		"detected":        lvr.Language,
		"expected":        lvr.ExpectedLang,
		"recommendation":  "Caption text should be in English (US) language",
	}
	
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return `{"type": "incorrect_language", "error": "Error marshalling JSON"}`
	}
	
	return string(jsonBytes)
}

// ValidateLanguage sends caption text to the language validation API
func ValidateLanguage(apiURL string, captionText string) (LanguageValidationResult, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create request with plaintext body
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(captionText))
	if err != nil {
		return LanguageValidationResult{}, fmt.Errorf("error creating request: %w", err)
	}
	
	// Set content type to plain text
	req.Header.Set("Content-Type", "text/plain")
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return LanguageValidationResult{}, fmt.Errorf("error sending request to language API: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return LanguageValidationResult{}, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, body)
	}
	
	// Parse response
	var langResp LanguageResponse
	if err := json.NewDecoder(resp.Body).Decode(&langResp); err != nil {
		return LanguageValidationResult{}, fmt.Errorf("error parsing API response: %w", err)
	}
	
	// Validate language is en-US
	const expectedLang = "en-US"
	isValid := langResp.Lang == expectedLang
	
	return LanguageValidationResult{
		Valid:        isValid,
		Type:         "incorrect_language",
		Language:     langResp.Lang,
		ExpectedLang: expectedLang,
	}, nil
}
