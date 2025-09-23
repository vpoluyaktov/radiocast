package charts

import (
	"strconv"
	"strings"
)

// parseFluxValue extracts numeric value from flux strings
func parseFluxValue(s string) float64 {
	// Remove whitespace and extract numeric part
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	
	// Try to parse as float
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	
	// Extract first number from string
	var result float64
	var current float64
	var hasNumber bool
	
	for _, char := range s {
		if char >= '0' && char <= '9' {
			current = current*10 + float64(char-'0')
			hasNumber = true
		} else if char == '.' && hasNumber {
			// Handle decimal point (simplified)
			break
		} else if hasNumber {
			break
		}
	}
	
	if hasNumber {
		result = current
	}
	
	return result
}

// getXrayValue converts X-ray flux string to numeric value for gauge
func getXrayValue(xray string) float64 {
	xray = strings.ToUpper(strings.TrimSpace(xray))
	if xray == "" {
		return 0
	}
	
	// Extract class and magnitude
	if len(xray) >= 2 {
		class := xray[0]
		magnitude := parseFluxValue(xray[1:])
		
		switch class {
		case 'A':
			return magnitude * 0.1 // A1 = 0.1, A9 = 0.9
		case 'B':
			return magnitude * 0.1 + 1 // B1 = 1.1, B9 = 1.9
		case 'C':
			return magnitude * 0.1 + 2 // C1 = 2.1, C9 = 2.9
		case 'M':
			return magnitude * 0.1 + 3 // M1 = 3.1, M9 = 3.9
		case 'X':
			return magnitude * 0.5 + 4 // X1 = 4.5, X9 = 8.5
		}
	}
	
	return 0
}
