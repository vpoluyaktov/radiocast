#!/bin/bash

# Radio Propagation Service - Unified Local Runner
# Consolidates all testing and debugging functionality into one script

# set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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

print_header() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

show_usage() {
    echo "Radio Propagation Service - Local Runner"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  server      Run complete local test (LLM + HTML generation)"
    echo "  debug-apis  Check all external API endpoints"
    echo "  unit-tests  Run Go unit tests"
    echo "  help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  OPENAI_API_KEY    Required - Your OpenAI API key"
    echo "  OPENAI_MODEL      Optional - Model to use (default: gpt-4.1)"
    echo "  PORT              Optional - Server port (default: 8981)"
    echo ""
    echo "Examples:"
    echo "  export OPENAI_API_KEY='sk-your-key-here'"
    echo "  $0 server"
}

check_requirements() {
    # Check if we're in the right directory
    if [ ! -f "main.go" ]; then
        print_error "Please run this script from the radiocast/service directory"
        exit 1
    fi

    # Check for Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check for OpenAI API key (except for unit tests and API debug)
    if [[ "$1" != "unit-tests" && "$1" != "debug-apis" && "$1" != "help" ]]; then
        if [ -z "$OPENAI_API_KEY" ]; then
            print_error "OPENAI_API_KEY environment variable is required"
            echo "Set it with: export OPENAI_API_KEY='sk-your-key-here'"
            exit 1
        fi
    fi
}

debug_apis() {
    print_header "API Endpoints Debug"
    
    print_status "Testing NOAA K-Index API..."
    if curl -s --max-time 10 "https://services.swpc.noaa.gov/json/planetary_k_index_1m.json" | head -3; then
        print_success "NOAA K-Index API responding"
    else
        print_error "NOAA K-Index API failed"
    fi
    
    echo ""
    print_status "Testing NOAA Solar API..."
    if curl -s --max-time 10 "https://services.swpc.noaa.gov/json/solar-cycle/observed-solar-cycle-indices.json" | head -3; then
        print_success "NOAA Solar API responding"
    else
        print_error "NOAA Solar API failed"
    fi
    
    echo ""
    print_status "Testing N0NBH API..."
    if curl -s --max-time 10 "https://www.hamqsl.com/solarapi.php?format=json" | head -3; then
        print_success "N0NBH API responding"
    else
        print_error "N0NBH API failed"
    fi
    
    echo ""
    print_status "Testing SIDC API..."
    if curl -s --max-time 10 "https://www.sidc.be/products/meu" | head -3; then
        print_success "SIDC API responding"
    else
        print_error "SIDC API failed"
    fi
}


run_server() {
    print_header "Complete Local Test"
    
    export ENVIRONMENT="local"
    export OPENAI_MODEL="${OPENAI_MODEL:-gpt-4.1}"
    export PORT="${PORT:-8981}"
    export LOCAL_REPORTS_DIR="./reports"
    
    print_status "Configuration:"
    print_status "  Environment: $ENVIRONMENT"
    print_status "  OpenAI Model: $OPENAI_MODEL"
    print_status "  Port: $PORT"
    print_status "  API Key: ${OPENAI_API_KEY:0:10}..."
    
    # Kill any existing process on the target port
    EXISTING_PID=$(lsof -ti:$PORT 2>/dev/null)
    if [ ! -z "$EXISTING_PID" ]; then
        print_status "🔄 Killing existing process on port $PORT (PID: $EXISTING_PID)"
        kill -TERM $EXISTING_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if kill -0 $EXISTING_PID 2>/dev/null; then
            kill -KILL $EXISTING_PID 2>/dev/null || true
            sleep 1
        fi
    fi
    
    # Clean up any existing reports
    rm -rf ./reports
    mkdir -p ./reports
    
    print_status "🧪 Testing complete pipeline..."
    print_status "  📡 Fetching real data from NOAA, N0NBH, and SIDC"
    print_status "  🤖 Generating report using OpenAI"
    print_status "  📊 Converting to HTML with charts"
    print_status "  ✅ Validating Chart Data and Band Analysis sections"
    
    # Start server briefly to generate a report
    go run main.go &
    SERVER_PID=$!
    sleep 3
    
    # Test health endpoint
    if curl -s http://localhost:$PORT/health | grep -q "healthy"; then
        print_success "✅ Server health check passed"
    else
        print_error "❌ Server health check failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
    
    # Generate report via HTTP
    REPORT_CONTENT=$(curl -s http://localhost:$PORT/)
    if [ $? -eq 0 ] && [ -n "$REPORT_CONTENT" ]; then
        kill $SERVER_PID 2>/dev/null || true
        
        # Validate report content
        if echo "$REPORT_CONTENT" | grep -q "Radio Propagation Report"; then
            print_success "✅ Report generated successfully"
        else
            print_error "❌ Report generation failed"
            return 1
        fi
        
        if echo "$REPORT_CONTENT" | grep -q "chart-container"; then
            print_success "✅ Charts found in HTML"
        else
            print_warning "⚠️  No charts found in HTML"
        fi
        
        if echo "$REPORT_CONTENT" | grep -q "Band-by-Band Analysis"; then
            if echo "$REPORT_CONTENT" | grep -q "band-analysis-table"; then
                print_success "✅ Band-by-Band Analysis table found"
            else
                print_warning "⚠️  Band-by-Band Analysis table found but missing CSS class"
            fi
        else
            print_warning "⚠️  Band-by-Band Analysis table missing"
        fi
        
        # Find the actual report file in the reports directory
        LATEST_REPORT_DIR=$(find ./reports -type d -name "????-??-??_??-??-??" | sort -r | head -1)
        if [ -n "$LATEST_REPORT_DIR" ] && [ -f "$LATEST_REPORT_DIR/index.html" ]; then
            FULL_PATH=$(realpath "$LATEST_REPORT_DIR/index.html")
            print_success "📄 Report saved: $LATEST_REPORT_DIR/index.html"
            print_success "🌐 Open in browser: file://$FULL_PATH"
        else
            print_warning "⚠️  Could not locate saved report file"
        fi
        
        
        print_success "🎉 Radio Propagation Service is working correctly!"
    else
        print_error "❌ Failed to fetch report from server"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
}


run_unit_tests() {
    print_header "Go Unit Tests"
    
    print_status "Running Go unit tests..."
    if go test -v ./...; then
        print_success "✅ All unit tests passed"
    else
        print_error "❌ Some unit tests failed"
        return 1
    fi
}

# Main script logic
case "${1:-help}" in
    "server")
        check_requirements "$1"
        run_server
        ;;
    "debug-apis")
        check_requirements "$1"
        debug_apis
        ;;
    "unit-tests")
        check_requirements "$1"
        run_unit_tests
        ;;
    "help"|*)
        show_usage
        ;;
esac
