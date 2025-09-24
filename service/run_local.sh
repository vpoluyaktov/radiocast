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

debug_helioviewer() {
    # Fetches the list of data sources from Helioviewer and tests the getClosestImage endpoint.
    print_status "Fetching Helioviewer data sources..."
    if ! curl -sS https://api.helioviewer.org/v2/getDataSources/ -o /tmp/helio_ds.json; then
        print_error "Failed to download Helioviewer data sources."
        return 1
    fi

    print_status "Downloaded Helioviewer data sources. First 5 lines:"
    head -n 5 /tmp/helio_ds.json

    SOURCE_ID=$(python3 -c '
import json, sys
try:
    with open("/tmp/helio_ds.json") as f:
        data = json.load(f)
    # The structure is [Observatory][Instrument][Detector][Measurement]
    sdo_aia = data.get("SDO", {}).get("AIA", {})
    if not sdo_aia:
        print("SDO/AIA not found in data sources", file=sys.stderr)
        sys.exit(1)

    for measurement in ["304", "193", "171"]:
        source_id = sdo_aia.get(measurement, {}).get("sourceId")
        if source_id:
            print(source_id)
            sys.exit(0)

    print("No sourceId found for SDO/AIA 304, 193, or 171", file=sys.stderr)
    sys.exit(1)
except Exception as e:
    print(f"Error parsing sourceId: {e}", file=sys.stderr)
    sys.exit(1)
'
    )

    if [[ -z "$SOURCE_ID" ]]; then
        print_error "Could not parse sourceId for SDO/AIA/304 from Helioviewer data sources."
        return 1
    fi
    print_success "Found sourceId for SDO/AIA 304: $SOURCE_ID"

    DATE_NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    print_status "Testing getClosestImage for date: $DATE_NOW with sourceId: $SOURCE_ID"

    if curl -sS "https://api.helioviewer.org/v2/getClosestImage/?date=$DATE_NOW&sourceId=$SOURCE_ID" | grep -q '"id"'; then
        print_success "Helioviewer getClosestImage API responding correctly."
    else
        print_error "Helioviewer getClosestImage API call failed."
        return 1
    fi
}

show_usage() {
    echo "Radio Propagation Service - Local Runner"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  server      Run complete local test (LLM + HTML generation)"
    echo "  stop        Stop running server process"
    echo "  debug-apis  Check all external API endpoints"
    echo "  unit-tests  Run Go unit tests"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  --mockup    Use mock data instead of real API calls (faster testing)"
    echo ""
    echo "Environment Variables:"
    echo "  OPENAI_API_KEY    Required - Your OpenAI API key"
    echo "  OPENAI_MODEL      Optional - Model to use (default: gpt-4.1)"
    echo "  PORT              Optional - Server port (default: 8981)"
    echo ""
    echo "Examples:"
    echo "  export OPENAI_API_KEY='sk-your-key-here'"
    echo "  $0 server"
    echo "  $0 server --mockup    # Fast testing with mock data"
    echo "  $0 stop               # Stop running server"
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
    if curl -s --max-time 10 "https://services.swpc.noaa.gov/products/noaa-planetary-k-index.json" | head -3; then
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

    echo ""
    print_status "Testing Helioviewer API..."
    debug_helioviewer
}


run_server() {
    local USE_MOCKUP=false
    
    # Check for --mockup flag
    if [[ "$2" == "--mockup" ]]; then
        USE_MOCKUP=true
    fi
    
    if [ "$USE_MOCKUP" = true ]; then
        print_header "Complete Local Test (MOCKUP MODE)"
        export MOCKUP_MODE=true
    else
        print_header "Complete Local Test"
    fi
    
    export ENVIRONMENT="local"
    export OPENAI_MODEL="${OPENAI_MODEL:-gpt-4.1}"
    export PORT="${PORT:-8981}"
    export LOCAL_REPORTS_DIR="./reports"
    
    print_status "Configuration:"
    print_status "  Environment: $ENVIRONMENT"
    if [ "$USE_MOCKUP" = true ]; then
        print_status "  Mode: MOCKUP (using mock data)"
    else
        print_status "  Mode: LIVE (real API calls)"
    fi
    print_status "  OpenAI Model: $OPENAI_MODEL"
    print_status "  Port: $PORT"
    print_status "  API Key: ${OPENAI_API_KEY:0:10}..."
    
    # Kill any existing process on the target port (safe method)
    EXISTING_PID=$(lsof -ti:$PORT 2>/dev/null)
    if [ ! -z "$EXISTING_PID" ]; then
        print_status "Killing existing process on port $PORT (PID: $EXISTING_PID)"
        kill -TERM $EXISTING_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if kill -0 $EXISTING_PID 2>/dev/null; then
            kill -KILL $EXISTING_PID 2>/dev/null || true
            sleep 1
        fi
    fi
    
    # Clean up any existing reports
    # rm -rf ./local_gcs
    # mkdir -p ./local_gcs
    
    print_status "Testing complete pipeline..."
    if [ "$USE_MOCKUP" = true ]; then
        print_status "  Using mock data from internal/mocks folder"
        print_status "  Loading pre-generated LLM response"
        print_status "  Using mock Sun GIF (no Helioviewer download)"
    else
        print_status "  Fetching real data from NOAA, N0NBH, and SIDC"
        print_status "  Generating report using OpenAI"
        print_status "  Downloading Sun images from Helioviewer"
    fi
    print_status "  Converting to HTML with charts"
    print_status "  Validating Chart Data and Band Analysis sections"
    
    # Start server briefly to generate a report
    go run main.go &
    SERVER_PID=$!
    sleep 3
    
    # Test health endpoint
    if curl -s http://localhost:$PORT/health | grep -q "healthy"; then
        print_success "Server health check passed"
    else
        print_error "Server health check failed"
        # Kill server process safely by PID
        if kill -0 $SERVER_PID 2>/dev/null; then
            kill -TERM $SERVER_PID 2>/dev/null || true
            sleep 1
        fi
        return 1
    fi
}

run_unit_tests() {
    print_header "Go Unit Tests"
    
    print_status "Running Go unit tests..."
    if go test -v ./...; then
        print_success "All unit tests passed"
    else
        print_error "Some unit tests failed"
        return 1
    fi
}

stop_server() {
    print_header "Stopping Radiocast Server"
    
    local PORT="${PORT:-8981}"
    
    print_status "Looking for process on port $PORT..."
    
    # Find process listening on the port
    EXISTING_PID=$(lsof -ti:$PORT 2>/dev/null)
    
    if [ -z "$EXISTING_PID" ]; then
        print_warning "No process found running on port $PORT"
        return 0
    fi
    
    print_status "Found process PID: $EXISTING_PID on port $PORT"
    
    # Get process details for confirmation
    PROCESS_INFO=$(ps -p $EXISTING_PID -o pid,ppid,cmd --no-headers 2>/dev/null)
    if [ ! -z "$PROCESS_INFO" ]; then
        print_status "Process details: $PROCESS_INFO"
    fi
    
    # Graceful termination first
    print_status "Sending TERM signal to process $EXISTING_PID..."
    if kill -TERM $EXISTING_PID 2>/dev/null; then
        print_status "TERM signal sent, waiting for graceful shutdown..."
        
        # Wait up to 5 seconds for graceful shutdown
        for i in {1..5}; do
            if ! kill -0 $EXISTING_PID 2>/dev/null; then
                print_success "Process $EXISTING_PID terminated gracefully"
                return 0
            fi
            sleep 1
            print_status "Waiting... ($i/5)"
        done
        
        # Force kill if still running
        if kill -0 $EXISTING_PID 2>/dev/null; then
            print_warning "Process still running, sending KILL signal..."
            if kill -KILL $EXISTING_PID 2>/dev/null; then
                sleep 1
                if ! kill -0 $EXISTING_PID 2>/dev/null; then
                    print_success "Process $EXISTING_PID force killed"
                else
                    print_error "Failed to kill process $EXISTING_PID"
                    return 1
                fi
            else
                print_error "Failed to send KILL signal to process $EXISTING_PID"
                return 1
            fi
        fi
    else
        print_error "Failed to send TERM signal to process $EXISTING_PID"
        return 1
    fi
}

# Main script logic
case "${1:-help}" in
    "server")
        check_requirements "$1"
        run_server "$1" "$2"
        ;;
    "stop")
        stop_server
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
