package parser

import (
	"strings"
	"testing"
)

func TestParseHTMLPreservesSpreadsheetTablesAndWorksheetHeadings(t *testing.T) {
	source := `<html><body><section><h2>Budget.html</h2><table><tr><th>Item</th><th>Price</th></tr><tr><td>Train</td><td>25</td></tr><tr><td>Hotel</td><td>120</td></tr></table></section><section><h2>Itinerary.html</h2><table><tr><td>Paris</td><td>Monday</td></tr><tr><td>Rome</td><td>Tuesday</td></tr></table></section></body></html>`
	parsed, err := NewDocumentParser().ParseHTML([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"## Budget.html", "| Item | Price |", "| Train | 25 |", "| Hotel | 120 |", "## Itinerary.html", "| Paris | Monday |", "| Rome | Tuesday |"} {
		if !strings.Contains(parsed.Content, fragment) {
			t.Fatalf("missing %q in markdown:\n%s", fragment, parsed.Content)
		}
	}
	if strings.Count(parsed.Content, "| ---") < 2 {
		t.Fatalf("worksheet tables were flattened:\n%s", parsed.Content)
	}
	if parsed.FileType != FileTypeHTML || parsed.WordCount == 0 {
		t.Fatal("parser metadata lost")
	}
}

func TestParseHTMLPreservesEmptyMultilineAndPipeCells(t *testing.T) {
	parsed, err := NewDocumentParser().ParseHTML([]byte(`<table><tr><th>Stop</th><th>Notes</th></tr><tr><td>Paris | France</td><td>Line one<br>Line two</td></tr><tr><td>Rome</td><td></td></tr></table>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`Paris \| France`, "Line one<br>Line two", "| Rome |  |"} {
		if !strings.Contains(parsed.Content, fragment) {
			t.Fatalf("cell content changed; expected %q in:\n%s", fragment, parsed.Content)
		}
	}
}
