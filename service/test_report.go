// Test runner for local report generation
// Usage: go run test_report.go local_runner.go test
package main

import (
	"log"
	"os"
)

func main() {
	// Check if we're running the local test
	if len(os.Args) > 1 && os.Args[1] == "test" {
		runLocalTest()
		return
	}
	
	log.Println("📋 Local Report Generation Test")
	log.Println("Usage: go run test_report.go local_runner.go test")
	log.Println("Make sure to set OPENAI_API_KEY environment variable")
	log.Println("")
	log.Println("Example:")
	log.Println("  export OPENAI_API_KEY='sk-your-key-here'")
	log.Println("  go run test_report.go local_runner.go test")
}
