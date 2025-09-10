package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Configuration options
var (
	port        = flag.String("port", "8080", "Port to listen on")
	language    = flag.String("language", "en-US", "Language code to return")
	forceDetect = flag.Bool("force-detect", false, "If true, detect language from text content")
)

func main() {
	flag.Parse()

	http.HandleFunc("/validate", handleValidate)
	http.HandleFunc("/", handleHome)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting mock language validation API on http://localhost%s", addr)
	log.Printf("Default language response: %s", *language)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<html>
			<head><title>Mock Language Validation API</title></head>
			<body>
				<h1>Mock Language API</h1>
				<p>This is a mock language validation API for testing the caption-validator.</p>
				<p>Current configuration:</p>
				<ul>
					<li>Default language response: %s</li>
					<li>Force detection: %t</li>
				</ul>
				<p>To test, send a POST request to /validate with plain text content.</p>
			</body>
		</html>
	`, *language, *forceDetect)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	text := string(body)

	// Log the request for debugging
	log.Printf("Received validation request: %s bytes of text", len(text))
	if len(text) > 100 {
		log.Printf("Text preview: %s...", text[:100])
	} else {
		log.Printf("Text: %s", text)
	}

	// Override default language detection based on content
	detectedLang := "en-US"
	
	// Check for French content
	frenchIndicators := []string{"bonjour", "merci", "comment", "allez-vous"}
	for _, word := range frenchIndicators {
		if strings.Contains(strings.ToLower(text), word) {
			log.Printf("French detected: found '%s'", word)
			detectedLang = "fr-FR"
			break
		}
	}
	
	// Check for Spanish content
	spanishIndicators := []string{"hola", "como", "está", "gracias", "por favor", "suscríbase"}
	for _, word := range spanishIndicators {
		if strings.Contains(strings.ToLower(text), word) {
			log.Printf("Spanish detected: found '%s'", word)
			detectedLang = "es-ES"
			break
		}
	}
	
	// Return the language response
	response := map[string]string{
		"lang": detectedLang,
	}
	
	w.Header().Set("Content-Type", "application/json")
	respBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	
	// Log the response for debugging
	log.Printf("Sending API response: %s", string(respBytes))
	
	// Write the response directly
	w.Write(respBytes)
}

// Simplified language detection for demo purposes
func detectLanguage(text string) string {
	text = strings.ToLower(text)
	
	// Simple non-English detection using common words
	nonEnglishIndicators := []string{
		// Spanish
		"hola", "como", "está", "gracias", "por favor", "buenos días",
		// French
		"bonjour", "merci", "comment", "vous", "français",
		// Other non-English indicators
		"schön", "danke", "ciao", "привет", "こんにちは",
	}
	
	// Check for any non-English words
	for _, word := range nonEnglishIndicators {
		if strings.Contains(text, word) {
			log.Printf("Detected non-English content: %s", word)
			// Just return a generic non-English indicator
			log.Printf("Language detection result: non-English")
			return "non-English"
		}
	}
	
	// Default to English US
	log.Printf("Language detection result: en-US (no non-English indicators found)")
	return "en-US"
}
