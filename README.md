# Caption-validator# Caption Validator

A command-line tool written in Go for validating WebVTT and SRT caption files.

## Features

- Supports WebVTT and SRT caption file formats
- Validates caption coverage percentage within a specified time range
- Validates caption language via an external API
- Outputs validation failures as JSON
- Outputs JSON error for unsupported file formats with exit code 1
- Clean error handling with no stack traces
- Supports extended time formats for TV shows and web series (2h, 30m, 1h30m)
- Provides batch processing for validating multiple caption files
- Memory-efficient parsing for large caption files
- Consistent JSON output format for all validation types

## Requirements

- Go 1.16+ (for building from source)
- Docker (for running containerized version)

## Installation

### Building from source

```bash
# Clone the repository
git clone <repository-url>
cd caption-validator

# Build the binary
go build -o caption-validator ./cmd/

# Run the binary
./caption-validator [flags] captions-file.vtt
```

**Note**: Make sure to use the `-o caption-validator ./cmd/` format when building to properly compile all files in the cmd directory.

### Using Docker

```bash
# Build the Docker image
docker build -t caption-validator .

# Run the container
docker run -v $(pwd):/data caption-validator [flags] /data/captions-file.vtt
```

#### Docker Examples

```bash
# Standard validation
docker run -v $(pwd):/data caption-validator -t_end 60 /data/captions.vtt

# Extended time format
docker run -v $(pwd):/data caption-validator -t_end 30m /data/episode.vtt

# With custom API for language validation
docker run -v $(pwd):/data caption-validator -t_end 1h -api https://api.example.com/lang /data/episode.vtt

# Batch processing with Docker
docker run -v $(pwd):/data -w /data caption-validator-batch ./batch-validate.sh ./episodes 30m 95
```

## Usage

```
caption-validator [flags] captions-filepath
```

### Important Note on Flag Format

Flags must be specified with a hyphen prefix. For example:

```
# Correct usage (with hyphens before flags)
caption-validator -t_end 30 -coverage 95 captions.vtt

# Incorrect usage (will not work)
caption-validator t_end 30 coverage 95 captions.vtt
```

### Flags

- `-coverage float`: Minimum percentage of time that should be covered by captions (default 95.0)
- `-t_start string`: Start time in seconds or HH:MM:SS format (default "0")
- `-t_end string`: End time in seconds or HH:MM:SS format (required)
- `-api string`: URL of the language validation API (default "http://localhost:8080/validate")

### Examples

#### Standard Validation

Validate a WebVTT file from 0 to 60 seconds with 95% coverage:
```bash
caption-validator -t_end 60 captions.vtt
```

Validate an SRT file from 00:01:30 to 00:05:00 with 90% coverage:
```bash
caption-validator -t_start 00:01:30 -t_end 00:05:00 -coverage 90 captions.srt
```

Use a custom language validation API:
```bash
caption-validator -t_end 60 -api https://api.example.com/lang captions.vtt
```

#### Extended Time Format Validation

Validate a 30-minute TV show episode:
```bash
# Using minutes notation
caption-validator -t_end 30m -coverage 95 episode.vtt

# Using hours and minutes notation
caption-validator -t_end 0h30m -coverage 95 episode.vtt
```

Validate a 1-hour TV show episode:
```bash
caption-validator -t_end 1h -coverage 95 episode.vtt
```

Validate a 1-hour-and-30-minute episode:
```bash
caption-validator -t_end 1h30m -coverage 95 episode.vtt
```

Validate a 2-hour episode with a custom start time:
```bash
caption-validator -t_start 5m -t_end 2h -coverage 95 episode.vtt
```

#### Batch Processing

The `batch-validate.sh` script allows you to validate multiple caption files in a directory:

```bash
# Basic batch validation (all caption files in directory)
./batch-validate.sh path/to/episodes 30m 95
```

Batch validation with language checking:
```bash
./batch-validate.sh path/to/episodes 30m 95 http://localhost:8080/validate
```

Batch validation with different time formats:
```bash
# Using seconds
./batch-validate.sh path/to/episodes 1800 95

# Using HH:MM:SS
./batch-validate.sh path/to/episodes 00:30:00 95

# Using hour notation
./batch-validate.sh path/to/episodes 0.5h 95
```

## Output

The program will output validation failures as JSON objects to stdout. If all validations pass, there will be no output. All validation errors use a consistent JSON format with a `type` field indicating the validation failure type.

### JSON Output Examples

#### 1. Caption Coverage Failure

```json
{"type": "caption_coverage", "required_coverage": 95, "actual_coverage": 85.75, "start_time": 0, "end_time": 60, "covered_time": 51.45, "total_time": 60, "missing_coverage_seconds": 5.55}
```

This indicates:
- The captions cover 85.75% of the specified time range (0 to 60 seconds)
- The required coverage was 95%
- 5.55 seconds of additional caption coverage would be needed to meet the requirement

#### 2. Language Validation Failure

```json
{"type": "incorrect_language", "detected": "es-ES", "expected": "en-US", "recommendation": "Caption text should be in English (US) language"}
```

This indicates:
- The language validation failed
- The caption text was detected as Spanish (es-ES)
- The expected language was English US (en-US)

#### 3. Unsupported Format Error

```json
{"type": "unsupported_format", "file": "./episodes/unsupported.txt", "error": "Unsupported caption file format"}
```

This indicates:
- The file format is not supported (neither WebVTT nor SRT)
- The program will exit with code 1 for this error

### Batch Processing Output

When using batch-validate.sh, the output is a JSON array with results for each file:

```json
[
  {"file": "./episodes/episode2.vtt", "results": {"actual_coverage":49.94,"covered_time":899,"end_time":1800,"missing_coverage_seconds":1,"required_coverage":50,"start_time":0,"total_time":1800,"type":"caption_coverage"}},
  {"file": "./episodes/episode4.vtt", "results": {"detected":"es-ES","expected":"en-US","recommendation":"Caption text should be in English (US) language","type":"incorrect_language"}}
]
```

## Error Handling and Exit Codes

### Exit Codes

- `0`: Program ran successfully (even if validation failures occurred)
- `1`: Program encountered an error (invalid file format, file not found, etc.)

### Improved Error Handling

The caption validator implements robust error handling:

#### 1. File Format Errors

- Unsupported file formats are detected early and reported as JSON errors
- Format errors include clear file path information for easy troubleshooting
- Exit code 1 is returned for unsupported formats

#### 2. Validation Failures

- Validation failures (coverage, language) are treated as validation results, not errors
- Program continues execution and returns exit code 0
- Detailed JSON output provides information for resolving validation issues

#### 3. Graceful Error Recovery

- API connection failures are handled gracefully
- Language validation is skipped if API is unavailable
- Detailed logging for troubleshooting without stack traces

#### 4. Batch Processing Error Handling

- Batch processing continues even if individual files fail
- Summary of failures is provided at the end
- JSON array includes all validation results and errors

## Architecture

The caption validator is designed with a focus on modularity and extensibility:

- `cmd/`: Contains the main application entry point
- `internal/parser/`: Handles detection and parsing of different caption formats
- `internal/validator/`: Implements validation logic for captions
- `internal/client/`: Contains HTTP client for language validation

This structure allows for easy addition of new caption formats or validation types in the future.

## Testing

The project includes comprehensive testing suite with unit tests, integration tests, and performance benchmarks.

### Unit Tests

Run the unit tests with:

```bash
go test ./...
```

### Integration Tests

The integration tests validate the end-to-end functionality of both caption-validator and batch-validate.sh:

```bash
# Run the integration tests
cd test
./integration_test.sh
```

The integration tests cover:
- Coverage validation with different thresholds
- Language validation with mock API
- Extended time format handling
- Processing multiple caption files

### Performance Tests

The performance tests evaluate the parsing efficiency with different file sizes and formats:

```bash
# Run performance benchmarks
cd test
go test -bench=. -benchmem performance_test.go
```

The benchmarks compare:
- Standard vs. optimized parsing methods
- WebVTT vs. SRT format parsing efficiency
- Memory usage for different file sizes (from 100 to 50,000 captions)
- Parsing time for different file sizes

### Mock Language API

For testing language validation, a mock API is provided:

```bash
# Start the mock API
cd mock_language_api
go run main.go --force-detect
```

The mock API supports:
- Language detection based on content patterns
- Forced detection mode for testing
- Configurable default language
