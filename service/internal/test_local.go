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
	
	log.Println("Use: go run . test")
	log.Println("Make sure to set OPENAI_API_KEY environment variable")
}
