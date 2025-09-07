package charts

// ChartSnippet represents an embeddable go-echarts chart fragment.
// Div should contain a single root <div id="..." style="..."></div>
// Script should contain the <script>...</script> block that initializes the chart in that div.
// HTML contains the complete snippet with div + script combined for template substitution.
type ChartSnippet struct {
    ID     string
    Title  string
    Div    string
    Script string
    HTML   string
    Width  string
    Height string
}


// conditionToValue maps band condition text to a numeric bucket for heatmap coloring.
// Returns: 0 Closed, 1 Poor, 2 Fair, 3 Good, 4 Excellent.
func (cg *ChartGenerator) conditionToValue(cond string) int {
    switch normalize(cond) {
    case "closed":
        return 0
    case "poor":
        return 1
    case "fair":
        return 2
    case "good":
        return 3
    case "excellent":
        return 4
    default:
        return 0
    }
}

// normalize trims and lowercases a string.
func normalize(s string) string {
    // simple inline to avoid extra imports
    // trim spaces
    start, end := 0, len(s)
    for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') { start++ }
    for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') { end-- }
    // lowercase ASCII only (inputs are expected ascii)
    out := make([]byte, end-start)
    for i := start; i < end; i++ {
        b := s[i]
        if b >= 'A' && b <= 'Z' { b = b + 32 }
        out[i-start] = b
    }
    return string(out)
}
