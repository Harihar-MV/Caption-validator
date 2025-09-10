#!/bin/bash
# batch-validate.sh - Simple batch processor for caption-validator
# Outputs only JSON to stdout and sends all logs to validation_logs.txt

# Define log file
LOG_FILE="validation_logs.txt"

# Log function - writes to log file instead of stdout/stderr
log() {
    echo "$@" >> "$LOG_FILE"
}

# Check for required arguments
if [ "$#" -lt 3 ]; then
    log "Usage: $0 <directory> <end_time> <min_coverage> [api_url]"
    log "Example: $0 ./season1 3600 95 http://localhost:8080/validate"
    log "Time formats: seconds (3600), HH:MM:SS (01:00:00), or with suffixes (1h, 30m, 1h30m)"
    echo "Error: Missing required arguments" >&2
    exit 1
fi

# Parameters
DIR=$1
END_TIME_INPUT=$2
MIN_COVERAGE=$3
API_URL=$4
OUTPUT_FILE="validation_results.json"

# Convert time format to seconds
convert_time_to_seconds() {
    local time_str=$1
    local seconds=0
    
    # If it's already a number, return it directly
    if [[ $time_str =~ ^[0-9]+(\.)?(\[0-9]+)?$ ]]; then
        echo $time_str
        return
    fi
    
    # Check for HH:MM:SS format
    if [[ $time_str =~ ^[0-9]+:[0-9]+:[0-9]+(\.[0-9]+)?$ ]]; then
        local hours=$(echo $time_str | cut -d":" -f1)
        local minutes=$(echo $time_str | cut -d":" -f2)
        local seconds=$(echo $time_str | cut -d":" -f3)
        echo $(($hours*3600 + $minutes*60 + $seconds))
        return
    fi
    
    # Check for MM:SS format
    if [[ $time_str =~ ^[0-9]+:[0-9]+(\.[0-9]+)?$ ]]; then
        local minutes=$(echo $time_str | cut -d":" -f1)
        local seconds=$(echo $time_str | cut -d":" -f2)
        echo $(($minutes*60 + $seconds))
        return
    fi
    
    # Check for hour suffix
    if [[ $time_str =~ h ]]; then
        local hours=$(echo $time_str | sed 's/h.*//')
        seconds=$(($seconds + $hours*3600))
        time_str=$(echo $time_str | sed 's/[0-9]*h//')
    fi
    
    # Check for minute suffix
    if [[ $time_str =~ m ]]; then
        local minutes=$(echo $time_str | sed 's/m.*//')
        seconds=$(($seconds + $minutes*60))
        time_str=$(echo $time_str | sed 's/[0-9]*m//')
    fi
    
    # Check for second suffix
    if [[ $time_str =~ s ]]; then
        local secs=$(echo $time_str | sed 's/s.*//')
        seconds=$(($seconds + $secs))
    fi
    
    echo $seconds
}

# Initialize log file
> "$LOG_FILE"
log "Starting batch validation at $(date)"

# Convert end time to seconds
END_TIME=$(convert_time_to_seconds "$END_TIME_INPUT")
log "Converted end time $END_TIME_INPUT to $END_TIME seconds"

# Initialize results array
echo "[" > "$OUTPUT_FILE"
COUNTER=0
UNSUPPORTED_FILES=0

# Check for unsupported file formats in the directory
for file in "$DIR"/*; do
    if [ -f "$file" ]; then
        ext="${file##*.}"
        if [ "$ext" != "vtt" ] && [ "$ext" != "srt" ] && [ "$ext" != "" ]; then
            log "Unsupported file format detected: $file"
            # Output JSON format error for unsupported files
            echo "{\"type\": \"unsupported_format\", \"file\": \"$file\", \"error\": \"Unsupported caption file format\"}"
            UNSUPPORTED_FILES=1
        fi
    fi
done

# We'll use a temporary file for logs instead of a pipe to avoid blocking
API_LOG_TMP="/tmp/api_log_tmp_$$"
> "$API_LOG_TMP"

# Process each caption file
for file in "$DIR"/*.vtt "$DIR"/*.srt; do
    if [ -f "$file" ]; then
        log "Processing $file..."
        
        # Run validation (capturing JSON errors to stdout, redirecting logs to log file)
        if [ -z "$API_URL" ]; then
            # Without API
            RESULT=$(./caption-validator -t_end "$END_TIME" -coverage "$MIN_COVERAGE" "$file" 2>> "$LOG_FILE")
        else
            # With API - capturing all logs to log file
            RESULT=$(./caption-validator -t_end "$END_TIME" -coverage "$MIN_COVERAGE" -api "$API_URL" "$file" 2> "$API_LOG_TMP")
        fi
        
        # If we have results, add to JSON
        if [ ! -z "$RESULT" ]; then
            if [ "$COUNTER" -gt 0 ]; then
                echo "," >> "$OUTPUT_FILE"
            fi
            echo "  {\"file\": \"$file\", \"results\": $RESULT}" >> "$OUTPUT_FILE"
            COUNTER=$((COUNTER+1))
        fi
    fi
done

# Close JSON array
echo "]" >> "$OUTPUT_FILE"

# Append any API logs to the main log file and clean up
if [ -f "$API_LOG_TMP" ]; then
    cat "$API_LOG_TMP" >> "$LOG_FILE"
    rm -f "$API_LOG_TMP"
fi

# Write summary to log file
log "Validation complete. Found $COUNTER issues."
log "Results saved to $OUTPUT_FILE"

# Output only the JSON to stdout
cat "$OUTPUT_FILE"

# Return exit code 1 if unsupported files were found
if [ $UNSUPPORTED_FILES -eq 1 ]; then
    exit 1
fi

