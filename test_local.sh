#!/bin/bash

# Radio Propagation Service - Local Test Script
# This script runs the report generation locally without GCS deployment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "service/main.go" ]; then
    print_error "Please run this script from the radiocast project root directory"
    exit 1
fi

# Check for OpenAI API key
if [ -z "$OPENAI_API_KEY" ]; then
    print_error "OPENAI_API_KEY environment variable is required"
    echo "Set it with: export OPENAI_API_KEY='sk-your-key-here'"
    exit 1
fi

print_status "🚀 Starting Radio Propagation Service Local Test"
print_status "OpenAI API Key: ${OPENAI_API_KEY:0:10}..."

# Change to service directory
cd service

# Set optional environment variables
export PORT="${PORT:-8080}"
export ENVIRONMENT="local"
export OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}"

print_status "📋 Configuration:"
print_status "  Port: $PORT"
print_status "  Environment: $ENVIRONMENT" 
print_status "  OpenAI Model: $OPENAI_MODEL"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "🔧 Go version: $(go version)"

# Run the local test
print_status "🧪 Running local report generation test..."
print_status "This will:"
print_status "  📡 Fetch real data from NOAA, N0NBH, and SIDC"
print_status "  🤖 Generate report using OpenAI"
print_status "  📊 Convert to HTML with charts"
print_status "  💾 Save as local HTML file"

echo ""
print_status "Starting test execution..."

# Run the test
if go run test_report.go local_runner.go test; then
    print_success "✅ Local test completed successfully!"
    
    # Find the generated HTML file
    LATEST_REPORT=$(ls -t test_report_*.html 2>/dev/null | head -1)
    if [ -n "$LATEST_REPORT" ]; then
        print_success "📄 Report saved as: $LATEST_REPORT"
        
        # Get absolute path
        FULL_PATH=$(realpath "$LATEST_REPORT")
        print_success "🌐 Open in browser: file://$FULL_PATH"
        
        # Try to open in browser (works on most systems)
        if command -v xdg-open &> /dev/null; then
            print_status "🚀 Opening report in default browser..."
            xdg-open "$FULL_PATH" &
        elif command -v open &> /dev/null; then
            print_status "🚀 Opening report in default browser..."
            open "$FULL_PATH" &
        else
            print_warning "Could not auto-open browser. Manually open: file://$FULL_PATH"
        fi
    fi
else
    print_error "❌ Local test failed"
    exit 1
fi

echo ""
print_success "🎉 Test completed! Check the generated HTML file for your propagation report."
