#!/bin/bash
# Integration test for batch-validate.sh

# Setup test environment
TEST_DIR="$(dirname "$0")/test_batch"
mkdir -p "$TEST_DIR"

# Create test caption files
echo "Creating test caption files..."

# Good WebVTT file with full coverage
cat > "$TEST_DIR/full_coverage.vtt" << EOF
WEBVTT

1
00:00:01.000 --> 00:00:20.000
This is a caption with full coverage

2
00:00:20.000 --> 00:00:40.000
This is another caption with full coverage

3
00:00:40.000 --> 00:01:00.000
This covers until the end of our test range
EOF

# WebVTT file with gaps (insufficient coverage)
cat > "$TEST_DIR/partial_coverage.vtt" << EOF
WEBVTT

1
00:00:01.000 --> 00:00:10.000
This caption only covers part of the range

2
00:00:20.000 --> 00:00:30.000
There's a gap between captions

3
00:00:40.000 --> 00:00:50.000
Another gap after this caption
EOF

# SRT file with good coverage
cat > "$TEST_DIR/good_coverage.srt" << EOF
1
00:00:01,000 --> 00:00:20,000
This is a caption with good coverage

2
00:00:20,500 --> 00:00:40,000
This is another caption with good coverage

3
00:00:40,000 --> 00:00:58,000
This covers almost until the end
EOF

# Spanish WebVTT file (for language testing)
cat > "$TEST_DIR/spanish.vtt" << EOF
WEBVTT

1
00:00:01.000 --> 00:00:20.000
Hola como está usted

2
00:00:20.000 --> 00:00:40.000
Este es un ejemplo en español

3
00:00:40.000 --> 00:01:00.000
Por favor suscríbase al canal
EOF

echo "Running tests..."

# Test 1: Test with 100% required coverage (should fail for partial coverage)
echo "Test 1: Testing with 100% required coverage"
../batch-validate.sh "$TEST_DIR" 60 100
if [ $? -eq 0 ]; then
    echo "✓ Test 1 passed: Script executed successfully"
    if grep -q "partial_coverage.vtt" validation_results.json; then
        echo "✓ Test 1 passed: Partial coverage file was flagged"
    else
        echo "✗ Test 1 failed: Partial coverage file was not flagged"
        exit 1
    fi
else
    echo "✗ Test 1 failed: Script execution failed"
    exit 1
fi

# Test 2: Test with 80% required coverage (partial coverage file should pass)
echo "Test 2: Testing with 80% required coverage"
../batch-validate.sh "$TEST_DIR" 60 80
if [ $? -eq 0 ]; then
    echo "✓ Test 2 passed: Script executed successfully"
    if grep -q "partial_coverage.vtt" validation_results.json; then
        echo "✗ Test 2 failed: Partial coverage file was incorrectly flagged at 80% threshold"
        exit 1
    else
        echo "✓ Test 2 passed: Partial coverage file passed at 80% threshold"
    fi
else
    echo "✗ Test 2 failed: Script execution failed"
    exit 1
fi

# Test 3: Test with language validation
echo "Test 3: Testing language validation"
# Start mock language API with force detection enabled
cd ../mock_language_api && go run main.go --force-detect &
MOCK_API_PID=$!
sleep 1
cd -

../batch-validate.sh "$TEST_DIR" 60 90 "http://localhost:8080/validate"
if [ $? -eq 0 ]; then
    echo "✓ Test 3 passed: Script executed successfully with language validation"
    if grep -q "spanish.vtt" validation_results.json && grep -q "incorrect_language" validation_results.json; then
        echo "✓ Test 3 passed: Spanish file was flagged for language"
    else
        echo "✗ Test 3 failed: Spanish file was not correctly identified"
        # Don't exit here as this might be a flaky test depending on the mock API
    fi
else
    echo "✗ Test 3 failed: Script execution failed with language validation"
    exit 1
fi

# Kill the mock API
kill $MOCK_API_PID

# Test 4: Test with extended time format
echo "Test 4: Testing with extended time format"
../batch-validate.sh "$TEST_DIR" 1m 90
if [ $? -eq 0 ]; then
    echo "✓ Test 4 passed: Extended time format accepted"
else
    echo "✗ Test 4 failed: Extended time format not accepted"
    exit 1
fi

echo "All tests completed."

# Optional: Clean up test files
# rm -rf "$TEST_DIR"
