package charts

import (
	"testing"
)

func TestChartSnippet(t *testing.T) {
	// Test ChartSnippet struct creation and field access
	snippet := ChartSnippet{
		ID:     "test-chart",
		Title:  "Test Chart",
		Div:    "<div id=\"test-chart\" style=\"width:100%;height:400px;\"></div>",
		Script: "<script>console.log('test');</script>",
		HTML:   "<div>Complete HTML</div>",
		Width:  "100%",
		Height: "400px",
	}
	
	if snippet.ID != "test-chart" {
		t.Errorf("Expected ID 'test-chart', got '%s'", snippet.ID)
	}
	if snippet.Title != "Test Chart" {
		t.Errorf("Expected Title 'Test Chart', got '%s'", snippet.Title)
	}
	if snippet.Width != "100%" {
		t.Errorf("Expected Width '100%%', got '%s'", snippet.Width)
	}
	if snippet.Height != "400px" {
		t.Errorf("Expected Height '400px', got '%s'", snippet.Height)
	}
}

func TestConditionToValue(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	tests := []struct {
		condition string
		expected  int
	}{
		{"Closed", 0},
		{"closed", 0},
		{"CLOSED", 0},
		{"Poor", 1},
		{"poor", 1},
		{"POOR", 1},
		{"Fair", 2},
		{"fair", 2},
		{"FAIR", 2},
		{"Good", 3},
		{"good", 3},
		{"GOOD", 3},
		{"Excellent", 4},
		{"excellent", 4},
		{"EXCELLENT", 4},
		{"Unknown", 0}, // Default case
		{"", 0},        // Empty string
		{"Invalid", 0}, // Invalid condition
	}
	
	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			result := generator.conditionToValue(tt.condition)
			if result != tt.expected {
				t.Errorf("conditionToValue(%s) = %d, expected %d", tt.condition, result, tt.expected)
			}
		})
	}
}

func TestConditionToValueWithWhitespace(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	tests := []struct {
		condition string
		expected  int
	}{
		{" Good ", 3},
		{"\tExcellent\t", 4},
		{"\nPoor\n", 1},
		{"\r\nFair\r\n", 2},
		{"  closed  ", 0},
		{" \t\n\r good \r\n\t ", 3},
	}
	
	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			result := generator.conditionToValue(tt.condition)
			if result != tt.expected {
				t.Errorf("conditionToValue(%q) = %d, expected %d", tt.condition, result, tt.expected)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Good", "good"},
		{"EXCELLENT", "excellent"},
		{"Fair", "fair"},
		{"poor", "poor"},
		{"Closed", "closed"},
		{"", ""},
		{"MiXeD CaSe", "mixed case"},
		{"123", "123"},
		{"Test123", "test123"},
		{"UPPER", "upper"},
		{"lower", "lower"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeWithWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{" Good ", "good"},
		{"\tExcellent\t", "excellent"},
		{"\nPoor\n", "poor"},
		{"\r\nFair\r\n", "fair"},
		{"  Closed  ", "closed"},
		{" \t\n\r Test \r\n\t ", "test"},
		{"   ", ""},
		{"\t\n\r", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeSpecialCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test-123", "test-123"},
		{"Test_456", "test_456"},
		{"Test.789", "test.789"},
		{"Test@Symbol", "test@symbol"},
		{"Test#Hash", "test#hash"},
		{"Test$Dollar", "test$dollar"},
		{"Test%Percent", "test%percent"},
		{"Test^Caret", "test^caret"},
		{"Test&Ampersand", "test&ampersand"},
		{"Test*Star", "test*star"},
		{"Test(Paren)", "test(paren)"},
		{"Test[Bracket]", "test[bracket]"},
		{"Test{Brace}", "test{brace}"},
		{"Test|Pipe", "test|pipe"},
		{"Test\\Backslash", "test\\backslash"},
		{"Test/Slash", "test/slash"},
		{"Test?Question", "test?question"},
		{"Test<Less>", "test<less>"},
		{"Test=Equal", "test=equal"},
		{"Test+Plus", "test+plus"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeEmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{" ", ""},
		{"\t", ""},
		{"\n", ""},
		{"\r", ""},
		{"\r\n", ""},
		{" \t\n\r ", ""},
		{"A", "a"},
		{"Z", "z"},
		{"a", "a"},
		{"z", "z"},
		{"0", "0"},
		{"9", "9"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConditionToValueConsistency(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Test that the same condition always returns the same value
	conditions := []string{"Good", "Fair", "Poor", "Excellent", "Closed"}
	
	for _, condition := range conditions {
		value1 := generator.conditionToValue(condition)
		value2 := generator.conditionToValue(condition)
		
		if value1 != value2 {
			t.Errorf("conditionToValue(%s) inconsistent: %d != %d", condition, value1, value2)
		}
		
		// Test case variations
		lowerValue := generator.conditionToValue(normalize(condition))
		upperValue := generator.conditionToValue(condition)
		
		if lowerValue != upperValue {
			t.Errorf("conditionToValue case sensitivity issue for %s: %d != %d", condition, lowerValue, upperValue)
		}
	}
}

func TestConditionToValueRange(t *testing.T) {
	generator := NewChartGenerator("/test")
	
	// Test that all valid conditions return values in expected range [0, 4]
	conditions := []string{"Closed", "Poor", "Fair", "Good", "Excellent"}
	expectedValues := []int{0, 1, 2, 3, 4}
	
	for i, condition := range conditions {
		value := generator.conditionToValue(condition)
		if value != expectedValues[i] {
			t.Errorf("conditionToValue(%s) = %d, expected %d", condition, value, expectedValues[i])
		}
		
		// Verify range
		if value < 0 || value > 4 {
			t.Errorf("conditionToValue(%s) = %d, out of expected range [0, 4]", condition, value)
		}
	}
}

func BenchmarkConditionToValue(b *testing.B) {
	generator := NewChartGenerator("/test")
	conditions := []string{"Good", "Fair", "Poor", "Excellent", "Closed", "Unknown"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		condition := conditions[i%len(conditions)]
		generator.conditionToValue(condition)
	}
}

func BenchmarkNormalize(b *testing.B) {
	inputs := []string{"Good", "EXCELLENT", "Fair", "poor", "Closed", "MiXeD CaSe"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		normalize(input)
	}
}
