package googledrive

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type exportTransport func(*http.Request) (*http.Response, error)

func (transport exportTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type archivePage struct{ name, content string }

func exportArchive(t *testing.T, pages ...archivePage) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, page := range pages {
		file, err := writer.Create(page.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(file, page.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
func fakeExportService(t *testing.T, mime string, data []byte, metadataStatus, exportStatus int) (*drive.Service, *[]string) {
	t.Helper()
	requests := []string{}
	transport := exportTransport(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.URL.String())
		if request.Method != "GET" {
			t.Errorf("unexpected method %s", request.Method)
		}
		if strings.HasSuffix(request.URL.Path, "/export") {
			want := "text/html"
			if mime == "application/vnd.google-apps.spreadsheet" {
				want = "application/zip"
			}
			if got := request.URL.Query().Get("mimeType"); got != want {
				t.Errorf("wrong MIME: got %q want %q", got, want)
			}
			return &http.Response{StatusCode: exportStatus, Header: http.Header{"Content-Type": []string{want}}, Body: io.NopCloser(bytes.NewReader(data)), Request: request}, nil
		}
		if request.URL.Query().Get("fields") != "mimeType" {
			t.Errorf("metadata fields must be minimal: %s", request.URL.RawQuery)
		}
		body, _ := json.Marshal(map[string]string{"mimeType": mime})
		return &http.Response{StatusCode: metadataStatus, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(body)), Request: request}, nil
	})
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(&http.Client{Transport: transport}), option.WithEndpoint("https://drive.example.test/"))
	if err != nil {
		t.Fatal(err)
	}
	return service, &requests
}

func TestHTMLExportUsesPlainHTMLForDocs(t *testing.T) {
	data := []byte("<html><body><h1>Itinerary</h1><p>Day one</p></body></html>")
	service, requests := fakeExportService(t, "application/vnd.google-apps.document", data, 200, 200)
	result, err := exportFileAsHTML(context.Background(), service, "document-id")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(result, data) || len(*requests) != 2 {
		t.Fatalf("document HTML changed: %s", result)
	}
}

func TestHTMLExportReadsEveryWorksheetAndPreservesNamesInOrder(t *testing.T) {
	data := exportArchive(t,
		archivePage{"Budget.html", "<!doctype html><html><head><title>Old title</title><style>hidden-css</style></head><body><h1>Costs</h1><table><tr><th>Item</th><th>Price</th></tr><tr><td>Train</td><td>25</td></tr></table></body></html>"},
		archivePage{"assets/style.css", "ignored asset"},
		archivePage{"tabs/<Places & dates>.HTML", "<html><body><table><tr><td>Paris</td><td>Monday</td></tr></table></body></html>"},
	)
	service, requests := fakeExportService(t, "application/vnd.google-apps.spreadsheet", data, 200, 200)
	result, err := exportFileAsHTML(context.Background(), service, "sheet-id")
	if err != nil {
		t.Fatal(err)
	}
	if len(*requests) != 2 {
		t.Fatal("unexpected export request count")
	}
	html := string(result)
	for _, fragment := range []string{"<h2>Budget.html</h2>", "<h1>Costs</h1>", "<td>Train</td>", "<td>25</td>", "<h2>tabs/&lt;Places &amp; dates&gt;.HTML</h2>", "<td>Paris</td>"} {
		if !strings.Contains(html, fragment) {
			t.Fatalf("missing %q in %s", fragment, html)
		}
	}
	if strings.Index(html, "Budget.html") > strings.Index(html, "Places &amp;") {
		t.Fatal("worksheet order changed")
	}
	for _, excluded := range []string{"hidden-css", "ignored asset", "<h2>tabs/<Places"} {
		if strings.Contains(html, excluded) {
			t.Fatalf("unsafe or non-content metadata leaked: %s", excluded)
		}
	}
}

func TestHTMLExportRejectsBadArchivesAndInflationLimits(t *testing.T) {
	tooMany := []archivePage{{"Sheet.html", "<table></table>"}}
	for i := 0; i < maxHTMLArchiveEntries; i++ {
		tooMany = append(tooMany, archivePage{fmt.Sprintf("assets/%d.css", i), ""})
	}
	oversizePage := strings.Repeat("x", maxHTMLExportBytes+1)
	half := strings.Repeat("y", maxHTMLExportBytes/2+1)
	cases := []struct {
		name string
		data []byte
	}{
		{"not a zip", []byte("not a zip")},
		{"no worksheets", exportArchive(t, archivePage{"assets/style.css", "body{}"})},
		{"too many entries", exportArchive(t, tooMany...)},
		{"inflated page", exportArchive(t, archivePage{"large.html", oversizePage})},
		{"inflated total", exportArchive(t, archivePage{"one.html", half}, archivePage{"two.html", half})},
		{"compressed size", make([]byte, maxHTMLExportBytes+1)},
		// HTML escaping can expand the combined representation beyond source bytes.
		{"rendered total", exportArchive(t, archivePage{"expanded.html", "<p>" + strings.Repeat("&", maxHTMLExportBytes/3) + "</p>"})},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if result, err := combineHTMLArchive(item.data); err == nil || result != nil {
				t.Fatalf("bad archive accepted: bytes=%d err=%v", len(result), err)
			}
		})
	}
}

func TestHTMLExportRejectsCorruptWorksheetChecksum(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.CreateHeader(&zip.FileHeader{Name: "Sheet.html", Method: zip.Store})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, "<p>Train fares</p>"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data := buffer.Bytes()
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	offset, err := archive.File[0].DataOffset()
	if err != nil {
		t.Fatal(err)
	}
	data[offset+3] = 'X'
	if result, err := combineHTMLArchive(data); err == nil || result != nil {
		t.Fatal("corrupt worksheet accepted")
	}
}

func TestHTMLExportPropagatesProviderFailuresAndBoundsDocs(t *testing.T) {
	for _, item := range []struct {
		name, mime                   string
		metadataStatus, exportStatus int
		data                         []byte
	}{
		{"metadata denied", "application/vnd.google-apps.spreadsheet", 403, 200, nil},
		{"export denied", "application/vnd.google-apps.spreadsheet", 200, 403, []byte(`{"error":{"message":"denied"}}`)},
		{"unsupported source", "application/pdf", 200, 200, nil},
		{"oversize document", "application/vnd.google-apps.document", 200, 200, make([]byte, maxHTMLExportBytes+1)},
		{"malformed sheet export", "application/vnd.google-apps.spreadsheet", 200, 200, []byte("<html>not an archive</html>")},
	} {
		t.Run(item.name, func(t *testing.T) {
			service, _ := fakeExportService(t, item.mime, item.data, item.metadataStatus, item.exportStatus)
			if result, err := exportFileAsHTML(context.Background(), service, "file-id"); err == nil || result != nil {
				t.Fatal("invalid provider export accepted")
			}
		})
	}
}
