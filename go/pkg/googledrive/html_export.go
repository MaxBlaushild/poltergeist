package googledrive

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"path"
	"strings"

	xhtml "golang.org/x/net/html"
	"google.golang.org/api/drive/v3"
)

const (
	maxHTMLExportBytes    = 10 * 1024 * 1024
	maxHTMLArchiveEntries = 100
)

// Sheets exports all worksheets as an HTML ZIP. Docs supports plain HTML.
// Keep the public Client method unchanged so both sources feed the same parser.
func exportFileAsHTML(ctx context.Context, service *drive.Service, fileID string) ([]byte, error) {
	file, err := service.Files.Get(fileID).Fields("mimeType").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect export file: %w", err)
	}
	mimeType := "text/html"
	switch file.MimeType {
	case "application/vnd.google-apps.document":
	case "application/vnd.google-apps.spreadsheet":
		mimeType = "application/zip"
	default:
		return nil, fmt.Errorf("HTML import supports Google Docs and Sheets only")
	}
	response, err := service.Files.Export(fileID, mimeType).Context(ctx).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to export file as HTML: %w", err)
	}
	defer response.Body.Close()
	data, err := readExportBounded(response.Body, maxHTMLExportBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read exported file: %w", err)
	}
	if mimeType == "text/html" {
		return data, nil
	}
	return combineHTMLArchive(data)
}

func readExportBounded(reader io.Reader, limit int) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("HTML export exceeds the %d-byte import limit", maxHTMLExportBytes)
	}
	return data, nil
}

type boundedHTMLBuffer struct{ buffer bytes.Buffer }

func (buffer *boundedHTMLBuffer) Len() int      { return buffer.buffer.Len() }
func (buffer *boundedHTMLBuffer) Bytes() []byte { return buffer.buffer.Bytes() }

func (buffer *boundedHTMLBuffer) Write(data []byte) (int, error) {
	if len(data) > maxHTMLExportBytes-buffer.Len() {
		return 0, fmt.Errorf("combined HTML export exceeds the %d-byte import limit", maxHTMLExportBytes)
	}
	return buffer.buffer.Write(data)
}
func (buffer *boundedHTMLBuffer) writeText(text string) error {
	_, err := buffer.Write([]byte(text))
	return err
}

// Nothing is extracted to the filesystem. Non-HTML assets are never inflated,
// and archive count, compressed bytes, inflated HTML and final output are bounded.
func combineHTMLArchive(data []byte) ([]byte, error) {
	if len(data) > maxHTMLExportBytes {
		return nil, fmt.Errorf("HTML archive exceeds import limit")
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid spreadsheet HTML archive: %w", err)
	}
	if len(archive.File) > maxHTMLArchiveEntries {
		return nil, fmt.Errorf("spreadsheet export contains more than %d archive entries", maxHTMLArchiveEntries)
	}
	var combined boundedHTMLBuffer
	if err := combined.writeText("<!doctype html><html><body>"); err != nil {
		return nil, err
	}
	pages, inflated := 0, 0
	for _, entry := range archive.File {
		extension := strings.ToLower(path.Ext(entry.Name))
		if !entry.FileInfo().Mode().IsRegular() || (extension != ".html" && extension != ".htm") {
			continue
		}
		remaining := maxHTMLExportBytes - inflated
		if entry.UncompressedSize64 > uint64(remaining) {
			return nil, fmt.Errorf("spreadsheet HTML exceeds the %d-byte import limit", maxHTMLExportBytes)
		}
		reader, err := entry.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot read spreadsheet HTML page: %w", err)
		}
		page, readErr := readExportBounded(reader, remaining)
		closeErr := reader.Close()
		if readErr != nil {
			return nil, fmt.Errorf("invalid spreadsheet HTML page: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("cannot close spreadsheet HTML page: %w", closeErr)
		}
		inflated += len(page)
		parsed, err := xhtml.Parse(bytes.NewReader(page))
		if err != nil {
			return nil, fmt.Errorf("invalid spreadsheet HTML: %w", err)
		}
		// Retain archive order and original page names as sheet headings. Escaping
		// filenames makes even unusual archive names inert text in the output.
		if err := combined.writeText("<section><h2>" + html.EscapeString(entry.Name) + "</h2>"); err != nil {
			return nil, err
		}
		body := htmlBody(parsed)
		if body != nil {
			for child := body.FirstChild; child != nil; child = child.NextSibling {
				if err := xhtml.Render(&combined, child); err != nil {
					return nil, err
				}
			}
		}
		if err := combined.writeText("</section>"); err != nil {
			return nil, err
		}
		pages++
	}
	if pages == 0 {
		return nil, fmt.Errorf("spreadsheet export contains no HTML worksheets")
	}
	if err := combined.writeText("</body></html>"); err != nil {
		return nil, err
	}
	return combined.Bytes(), nil
}

func htmlBody(node *xhtml.Node) *xhtml.Node {
	if node.Type == xhtml.ElementNode && node.Data == "body" {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := htmlBody(child); found != nil {
			return found
		}
	}
	return nil
}
