package parser

import (
	"bytes"
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/nguyenthenguyen/docx"
	"rsc.io/pdf"
)

type DocumentParser struct{}

type ParsedDocument struct {
	Content   string `json:"content"`
	FileType  string `json:"fileType"`
	PageCount *int   `json:"pageCount,omitempty"`
	WordCount int    `json:"wordCount"`
}

const (
	FileTypePDF  = "pdf"
	FileTypeDOCX = "docx"
	FileTypeHTML = "html"
	MaxFileSize  = 10 * 1024 * 1024 // 10MB
)

func NewDocumentParser() *DocumentParser {
	return &DocumentParser{}
}

// DetectFileType detects the file type from the file bytes
func (p *DocumentParser) DetectFileType(fileBytes []byte) (string, error) {
	if len(fileBytes) < 4 {
		return "", fmt.Errorf("file too small to detect type")
	}

	// PDF signature: %PDF
	if bytes.HasPrefix(fileBytes, []byte("%PDF")) {
		return FileTypePDF, nil
	}

	// DOCX is a ZIP file containing XML, check for ZIP signature (PK)
	// DOCX files start with PK\x03\x04 (ZIP file signature)
	if bytes.HasPrefix(fileBytes, []byte("PK\x03\x04")) {
		// Check if it's actually a docx by looking for specific file structure
		if bytes.Contains(fileBytes, []byte("word/")) {
			return FileTypeDOCX, nil
		}
		return "", fmt.Errorf("unsupported ZIP-based document format")
	}

	return "", fmt.Errorf("unsupported file type")
}

// ParsePDF extracts text from a PDF file
func (p *DocumentParser) ParsePDF(fileBytes []byte) (*ParsedDocument, error) {
	reader := bytes.NewReader(fileBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(fileBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF: %w", err)
	}

	numPages := pdfReader.NumPage()
	var allText strings.Builder

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		content := page.Content()
		for _, text := range content.Text {
			allText.WriteString(text.S)
			allText.WriteString(" ")
		}
	}

	content := strings.TrimSpace(allText.String())
	wordCount := len(strings.Fields(content))

	return &ParsedDocument{
		Content:   content,
		FileType:  FileTypePDF,
		PageCount: &numPages,
		WordCount: wordCount,
	}, nil
}

// ParseWord extracts text from a Word (.docx) file
func (p *DocumentParser) ParseWord(fileBytes []byte) (*ParsedDocument, error) {
	reader := bytes.NewReader(fileBytes)

	docxReader, err := docx.ReadDocxFromMemory(reader, int64(len(fileBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Word document: %w", err)
	}
	defer docxReader.Close()

	// Get the document content
	docxDoc := docxReader.Editable()
	content := docxDoc.GetContent()

	// Clean up the content - remove extra whitespace
	content = strings.TrimSpace(content)

	wordCount := len(strings.Fields(content))

	return &ParsedDocument{
		Content:   content,
		FileType:  FileTypeDOCX,
		PageCount: nil, // Word documents don't have a fixed page count
		WordCount: wordCount,
	}, nil
}

// ParseHTML converts HTML to markdown
func (p *DocumentParser) ParseHTML(htmlBytes []byte) (*ParsedDocument, error) {
	html := string(htmlBytes)

	// Convert HTML to markdown
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	// Clean up the content
	content := strings.TrimSpace(markdown)
	wordCount := len(strings.Fields(content))

	return &ParsedDocument{
		Content:   content,
		FileType:  FileTypeHTML,
		PageCount: nil, // HTML doesn't have a fixed page count
		WordCount: wordCount,
	}, nil
}

// ParseDocument detects file type and parses the document accordingly
func (p *DocumentParser) ParseDocument(fileBytes []byte) (*ParsedDocument, error) {
	// Check file size
	if len(fileBytes) > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// Detect file type
	fileType, err := p.DetectFileType(fileBytes)
	if err != nil {
		return nil, err
	}

	// Parse based on file type
	switch fileType {
	case FileTypePDF:
		return p.ParsePDF(fileBytes)
	case FileTypeDOCX:
		return p.ParseWord(fileBytes)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}
