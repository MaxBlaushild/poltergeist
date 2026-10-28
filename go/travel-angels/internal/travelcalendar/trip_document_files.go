package travelcalendar

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxTripDocumentBytes = 20 << 20

func tripDocumentFilename(filename string) string {
	filename = path.Base(strings.ReplaceAll(filename, "\\", "/"))
	filename = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, filename)
	filename = strings.TrimSpace(filename)
	if filename == "." || filename == "" || !utf8.ValidString(filename) || utf8.RuneCountInString(filename) > 180 {
		return ""
	}
	return filename
}

// Files are never fetched from a submitted URL or served as active HTML. The
// original is stored privately and downloaded through the trip's access check.
func validateTripDocumentFile(filename string, data []byte) (string, string, error) {
	invalid := bad("Upload a PDF, DOCX, XLSX, CSV, TXT, JPEG, or PNG file up to 20 MB")
	if len(data) == 0 || len(data) > maxTripDocumentBytes || filename == "" {
		return "", "", invalid
	}
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".pdf":
		if !bytes.HasPrefix(data, []byte("%PDF-")) {
			return "", "", invalid
		}
		return "application/pdf", "guide", nil
	case ".jpg", ".jpeg", ".png":
		cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 40_000_000 ||
			(ext == ".png" && kind != "png") || (ext != ".png" && kind != "jpeg") {
			return "", "", invalid
		}
		if kind == "png" {
			return "image/png", "guide", nil
		}
		return "image/jpeg", "guide", nil
	case ".txt", ".csv":
		if !utf8.Valid(data) {
			return "", "", invalid
		}
		for _, r := range string(data) {
			if unicode.IsControl(r) && r != '\r' && r != '\n' && r != '\t' {
				return "", "", invalid
			}
		}
		if strings.Contains(http.DetectContentType(data), "html") || strings.Contains(http.DetectContentType(data), "xml") {
			return "", "", invalid
		}
		if ext == ".csv" {
			return "text/csv; charset=utf-8", "spreadsheet", nil
		}
		return "text/plain; charset=utf-8", "guide", nil
	case ".docx", ".xlsx":
		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil || len(archive.File) > 2000 {
			return "", "", invalid
		}
		required := "word/document.xml"
		mimeType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		mainType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
		format := "guide"
		if ext == ".xlsx" {
			required = "xl/workbook.xml"
			mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			mainType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
			format = "spreadsheet"
		}
		var total uint64
		found, types := false, false
		for _, file := range archive.File {
			total += file.UncompressedSize64
			name := strings.ReplaceAll(file.Name, "\\", "/")
			if total > 100<<20 || file.UncompressedSize64 > 64<<20 || strings.HasPrefix(name, "/") || strings.Contains(name, "../") || strings.Contains(strings.ToLower(name), "vbaproject") {
				return "", "", invalid
			}
			if name == required && file.UncompressedSize64 > 0 {
				found = true
			}
			if name == "[Content_Types].xml" {
				if file.UncompressedSize64 > 1<<20 {
					return "", "", invalid
				}
				reader, err := file.Open()
				if err != nil {
					return "", "", invalid
				}
				content, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
				reader.Close()
				if err != nil || len(content) > 1<<20 || bytes.Contains(bytes.ToLower(content), []byte("macroenabled")) {
					return "", "", invalid
				}
				types = bytes.Contains(content, []byte(mainType))
			}
		}
		if !found || !types {
			return "", "", invalid
		}
		return mimeType, format, nil
	default:
		return "", "", invalid
	}
}

func (s *Service) deleteTripDocumentObject(document *TripDocument) {
	if document == nil || document.StorageKey == "" || s.tripImageStorage == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.tripImageStorage.DeleteObject(ctx, document.StorageBucket, document.StorageKey); err != nil {
		// Do not log filenames, content, or upstream errors that could contain
		// signed URLs. The document ID is enough to investigate orphan cleanup.
		log.Printf("trip document object cleanup failed for document %s", document.ID)
	}
}

func (s *Service) uploadTripDocument(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	viewer := s.viewer(c)
	viewer.UserID = user.ID
	// Authorize before reading a potentially large upload, then again after
	// storage completes. Never hold calendar permission locks across S3 IO.
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, _, err := s.tripDocumentAccess(tx, c, viewer, true)
		if err != nil {
			return err
		}
		return checkTripDocumentCapacity(tx, stop.ID)
	})
	if err != nil {
		respondError(c, err)
		return
	}
	if s.tripImageStorage == nil || s.tripImageBucket == "" {
		c.JSON(503, gin.H{"error": "Document uploads are temporarily unavailable. You can still share a link"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTripDocumentBytes+(64<<10))
	if err = c.Request.ParseMultipartForm(1 << 20); err != nil {
		respondError(c, bad("Upload one document no larger than 20 MB"))
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	form := c.Request.MultipartForm
	if form == nil || len(form.File) != 1 || len(form.File["file"]) != 1 {
		respondError(c, bad("Choose one document to upload"))
		return
	}
	header := form.File["file"][0]
	if header.Size <= 0 || header.Size > maxTripDocumentBytes {
		respondError(c, bad("Documents must be no larger than 20 MB"))
		return
	}
	file, err := header.Open()
	if err != nil {
		respondError(c, bad("Unable to read the uploaded document"))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxTripDocumentBytes+1))
	if err != nil {
		respondError(c, bad("Unable to read the uploaded document"))
		return
	}
	document := TripDocument{ID: uuid.New(), StopID: uuid.MustParse(c.Param("id")), AddedBy: user.ID, Kind: "file", Title: strings.TrimSpace(c.PostForm("title")), Filename: tripDocumentFilename(header.Filename), SizeBytes: int64(len(data)), CreatedAt: time.Now().UTC()}
	document.ContentType, document.Format, err = validateTripDocumentFile(document.Filename, data)
	if err != nil {
		respondError(c, err)
		return
	}
	if document.Title == "" {
		document.Title = document.Filename
	}
	if err = validateTripDocumentTitle(document.Title); err != nil {
		respondError(c, err)
		return
	}
	document.StorageBucket = s.tripImageBucket
	document.StorageKey = "trip-documents/" + document.StopID.String() + "/" + document.ID.String()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	err = s.tripImageStorage.PutObject(ctx, document.StorageBucket, document.StorageKey, document.ContentType, data)
	cancel()
	if err != nil {
		s.deleteTripDocumentObject(&document)
		respondError(c, err)
		return
	}
	var result gin.H
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, current, err := s.tripDocumentAccess(tx, c, viewer, true)
		if err != nil {
			return err
		}
		if err = checkTripDocumentCapacity(tx, stop.ID); err != nil {
			return err
		}
		if err = tx.Create(&document).Error; err != nil {
			return err
		}
		result, err = s.tripDocumentDTO(tx, stop, &document, current, true)
		return err
	})
	if err != nil {
		s.deleteTripDocumentObject(&document)
		respondError(c, err)
		return
	}
	c.JSON(201, gin.H{"document": result})
}

func (s *Service) getTripDocumentFile(c *gin.Context) {
	viewer := s.viewer(c)
	var document *TripDocument
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, _, err := s.tripDocumentAccess(tx, c, viewer, false)
		if err != nil {
			return err
		}
		document, err = loadTripDocument(tx, stop.ID, c.Param("documentId"))
		if err != nil {
			return err
		}
		if document.Kind != "file" || document.StorageKey == "" || s.tripImageStorage == nil {
			return errUnavailable
		}
		return nil
	})
	if err != nil {
		respondError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	data, err := s.tripImageStorage.ReadObject(ctx, document.StorageBucket, document.StorageKey, maxTripDocumentBytes)
	cancel()
	if err != nil {
		respondError(c, err)
		return
	}
	if int64(len(data)) != document.SizeBytes || len(data) == 0 || len(data) > maxTripDocumentBytes {
		respondError(c, errUnavailable)
		return
	}
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, _, err := s.tripDocumentAccess(tx, c, viewer, false)
		if err != nil {
			return err
		}
		if _, err = loadTripDocument(tx, stop.ID, document.ID.String()); err != nil {
			return err
		}
		c.Header("Cache-Control", "private, no-store, max-age=0")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Security-Policy", "sandbox; default-src 'none'")
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": document.Filename}))
		c.Data(200, document.ContentType, data)
		return nil
	})
	if err != nil {
		respondError(c, err)
	}
}
