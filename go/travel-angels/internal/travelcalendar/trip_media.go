package travelcalendar

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxTripImageBytes = 8 << 20
const maxTripImagePixels = 20_000_000
const maxTripImageSide = 2400

// TripImageStorage is implemented by the shared AWS file-storage client. The
// bucket must block public access; clients only receive the protected API URL.
type TripImageStorage interface {
	PutObject(context.Context, string, string, string, []byte) error
	ReadObject(context.Context, string, string, int64) ([]byte, error)
	DeleteObject(context.Context, string, string) error
}

func (s *Service) ConfigureTripImageStorage(storage TripImageStorage, bucket string) {
	s.tripImageStorage = storage
	s.tripImageBucket = strings.TrimSpace(bucket)
}

// The caller holds the trip row lock. Keep both versions until publication
// changes, so editing a cover never replaces the image guests currently see.
func pruneTripMedia(tx *gorm.DB, stop *Stop) ([]TripMedia, error) {
	keep := []uuid.UUID{}
	if stop.Draft.Hero != nil {
		keep = append(keep, stop.Draft.Hero.MediaID)
	}
	if stop.Published != nil && stop.Published.Hero != nil {
		keep = append(keep, stop.Published.Hero.MediaID)
	}
	for _, content := range []*StopContent{&stop.Draft, stop.Published} {
		if content == nil {
			continue
		}
		for _, lodging := range content.Lodging {
			for _, photo := range lodging.Photos {
				keep = append(keep, photo.MediaID)
			}
		}
	}
	query := tx.Where("stop_id = ?", stop.ID)
	if len(keep) > 0 {
		query = query.Where("id NOT IN ?", keep)
	}
	var obsolete []TripMedia
	if err := query.Select("id, storage_bucket, storage_key").Find(&obsolete).Error; err != nil {
		return nil, err
	}
	if err := query.Delete(&TripMedia{}).Error; err != nil {
		return nil, err
	}
	return obsolete, nil
}

func (s *Service) deleteTripImageObjects(media []TripMedia) {
	if s.tripImageStorage == nil {
		return
	}
	for _, item := range media {
		if item.StorageKey == "" {
			continue
		}
		// Cleanup must still run after the original request was cancelled.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.tripImageStorage.DeleteObject(ctx, item.StorageBucket, item.StorageKey)
		cancel()
		if err != nil {
			log.Printf("trip image object cleanup failed for media %s (s3://%s/%s): %v", item.ID, item.StorageBucket, item.StorageKey, err)
		}
	}
}

type TripHero struct {
	MediaID    uuid.UUID `json:"mediaId"`
	AltText    string    `json:"altText"`
	FocalX     float64   `json:"focalX"`
	FocalY     float64   `json:"focalY"`
	CropX      float64   `json:"cropX"`
	CropY      float64   `json:"cropY"`
	CropWidth  float64   `json:"cropWidth"`
	CropHeight float64   `json:"cropHeight"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	URL        string    `json:"url"`
}
type TripMedia struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	StopID        uuid.UUID  `gorm:"type:uuid;index"`
	LodgingID     *uuid.UUID `gorm:"type:uuid"`
	UploadedBy    uuid.UUID  `gorm:"type:uuid"`
	ContentType   string
	Width         int
	Height        int
	Data          []byte
	StorageBucket string
	StorageKey    string
	CreatedAt     time.Time
}

func (TripMedia) TableName() string { return "tc_trip_media" }
func (h *TripHero) validate() error {
	h.AltText = strings.TrimSpace(h.AltText)
	if h.MediaID == uuid.Nil || h.AltText == "" || utf8.RuneCountInString(h.AltText) > 500 {
		return bad("A hero image needs its uploaded media ID and an accessible description up to 500 characters")
	}
	if h.CropWidth == 0 && h.CropHeight == 0 {
		h.CropWidth = 1
		h.CropHeight = 1
	}
	for _, v := range []float64{h.FocalX, h.FocalY, h.CropX, h.CropY, h.CropWidth, h.CropHeight} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return bad("Image crop and focal coordinates must be between zero and one")
		}
	}
	if h.CropWidth <= 0 || h.CropHeight <= 0 || h.CropX+h.CropWidth > 1.0000001 || h.CropY+h.CropHeight > 1.0000001 {
		return bad("Keep the crop rectangle inside the image")
	}
	return nil
}
func tripHeroURL(stopID, mediaID uuid.UUID) string {
	return "/travel-angels/calendar/stops/" + stopID.String() + "/hero/" + mediaID.String()
}
func (s *Service) registerTripMediaRoutes(g *gin.RouterGroup) {
	s.registerLodgingMediaRoutes(g)
	g.POST("/calendar/stops/:id/hero", s.uploadTripHero)
	g.GET("/calendar/stops/:id/hero/:mediaId", s.getTripHero)
	g.PUT("/calendar/stops/:id/logistics/permissions", s.updateTripLogisticsPermissions)
}
func (s *Service) uploadTripHero(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTripImageBytes+(64<<10))
	if err := c.Request.ParseMultipartForm(maxTripImageBytes); err != nil {
		respondError(c, bad("Upload one JPEG or PNG image no larger than 8 MB"))
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if c.Request.MultipartForm == nil || len(c.Request.MultipartForm.File["image"]) != 1 {
		respondError(c, bad("Upload one JPEG or PNG image"))
		return
	}
	header := c.Request.MultipartForm.File["image"][0]
	if header.Size > maxTripImageBytes {
		respondError(c, bad("Hero images must be no larger than 8 MB"))
		return
	}
	file, err := header.Open()
	if err != nil {
		respondError(c, bad("Unable to read the uploaded image"))
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxTripImageBytes+1))
	if err != nil || len(raw) > maxTripImageBytes {
		respondError(c, bad("Unable to read an image within the 8 MB limit"))
		return
	}
	hero := TripHero{MediaID: uuid.New(), AltText: c.PostForm("altText"), FocalX: .5, FocalY: .5, CropWidth: 1, CropHeight: 1}
	for field, target := range map[string]*float64{"focalX": &hero.FocalX, "focalY": &hero.FocalY, "cropX": &hero.CropX, "cropY": &hero.CropY, "cropWidth": &hero.CropWidth, "cropHeight": &hero.CropHeight} {
		if value := c.PostForm(field); value != "" {
			parsed, e := strconv.ParseFloat(value, 64)
			if e != nil {
				respondError(c, bad("Invalid image crop coordinates"))
				return
			}
			*target = parsed
		}
	}
	if err = hero.validate(); err != nil {
		respondError(c, err)
		return
	}
	var revision *int
	if rawRevision := c.PostForm("revision"); rawRevision != "" {
		parsed, e := strconv.Atoi(rawRevision)
		if e != nil {
			respondError(c, bad("Invalid trip revision"))
			return
		}
		revision = &parsed
	}
	sanitized, contentType, width, height, err := sanitizeTripImage(raw)
	if err != nil {
		respondError(c, err)
		return
	}
	hero.Width = width
	hero.Height = height
	var stop *Stop
	var uploaded *TripMedia
	var obsolete []TripMedia
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var err error
		stop, err = s.lockedStop(tx, c.Param("id"))
		if err != nil {
			return err
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: user.ID}) {
			return errForbidden
		}
		if err = s.checkTripRevision(stop, revision); err != nil {
			return err
		}
		hero.URL = tripHeroURL(stop.ID, hero.MediaID)
		media := TripMedia{ID: hero.MediaID, StopID: stop.ID, UploadedBy: user.ID, ContentType: contentType, Width: width, Height: height, Data: sanitized, CreatedAt: time.Now().UTC()}
		if s.tripImageStorage != nil && s.tripImageBucket != "" {
			extension := ".jpg"
			if contentType == "image/png" {
				extension = ".png"
			}
			media.StorageBucket = s.tripImageBucket
			media.StorageKey = "trip-heroes/" + stop.ID.String() + "/" + hero.MediaID.String() + extension
			media.Data = []byte{}
			uploaded = &media
			ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
			err = s.tripImageStorage.PutObject(ctx, media.StorageBucket, media.StorageKey, contentType, sanitized)
			cancel()
			if err != nil {
				return err
			}
		}
		if err = tx.Create(&media).Error; err != nil {
			return err
		}
		stop.Draft.Hero = &hero
		if err = s.saveTripChange(tx, stop, user.ID, "hero_uploaded", gin.H{"mediaId": hero.MediaID}); err != nil {
			return err
		}
		obsolete, err = pruneTripMedia(tx, stop)
		return err
	})
	if err != nil {
		if uploaded != nil {
			s.deleteTripImageObjects([]TripMedia{*uploaded})
		}
		respondError(c, err)
		return
	}
	s.deleteTripImageObjects(obsolete)
	c.JSON(200, gin.H{"stop": s.stopDTO(stop, Viewer{UserID: user.ID})})
}
func (s *Service) currentTripMediaViewer(tx *gorm.DB, stop *Stop, c *gin.Context, viewer Viewer) Viewer {
	// Link capabilities must still match now, after acquiring permission locks.
	viewer.CalendarLinkID = uuid.Nil
	viewer.StopLinkID = uuid.Nil
	if viewer.UserID != uuid.Nil {
		stored := s.participationViewer(tx, stop, viewer.UserID, participationInput{})
		viewer.CalendarLinkID = stored.CalendarLinkID
		viewer.StopLinkID = stored.StopLinkID
	}
	if stop.SharingEnabled && c.Query("stopToken") != "" && c.Query("stopToken") == stop.ShareToken {
		viewer.StopLinkID = stop.ID
	}
	if token := c.Query("calendarToken"); token != "" {
		var calendar Calendar
		if tx.Where("id = ? AND share_token = ? AND sharing_enabled = true", stop.CalendarID, token).First(&calendar).Error == nil {
			viewer.CalendarLinkID = calendar.ID
		}
	}
	return viewer
}

func (s *Service) authorizeTripHero(tx *gorm.DB, c *gin.Context, viewer Viewer, mediaID uuid.UUID) error {
	stop, err := s.lockedStop(tx, c.Param("id"))
	if err != nil {
		return err
	}
	viewer = s.currentTripMediaViewer(tx, stop, c, viewer)
	if !s.canViewStop(tx, stop, viewer) {
		return errUnavailable
	}
	manager := s.canManageStop(tx, stop, viewer)
	attached := stop.Published != nil && stop.Published.Hero != nil && stop.Published.Hero.MediaID == mediaID
	if manager && stop.Draft.Hero != nil && stop.Draft.Hero.MediaID == mediaID {
		attached = true
	}
	if !attached {
		return errUnavailable
	}
	return nil
}

func (s *Service) getTripHero(c *gin.Context) {
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		respondError(c, errUnavailable)
		return
	}
	requestedWidth := 0
	if requested := c.Query("width"); requested != "" {
		requestedWidth, err = strconv.Atoi(requested)
		if err != nil || requestedWidth < 64 || requestedWidth > maxTripImageSide {
			respondError(c, bad("Choose an image width between 64 and 2400 pixels"))
			return
		}
	}
	// Authentication can use another DB connection, so resolve it before locks.
	viewer := s.viewer(c)
	var media TripMedia
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := s.authorizeTripHero(tx, c, viewer, mediaID); err != nil {
			return err
		}
		return tx.Where("id = ? AND stop_id = ? AND lodging_id IS NULL", mediaID, c.Param("id")).First(&media).Error
	})
	if err != nil {
		respondError(c, err)
		return
	}
	// Network reads and image processing must not hold the owning calendar's
	// permission lock, which would block unrelated trips and RSVP changes.
	data := media.Data
	if media.StorageKey != "" {
		if s.tripImageStorage == nil {
			respondError(c, errUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		data, err = s.tripImageStorage.ReadObject(ctx, media.StorageBucket, media.StorageKey, maxTripImageBytes)
		cancel()
		if err != nil {
			respondError(c, err)
			return
		}
		if len(data) == 0 || len(data) > maxTripImageBytes {
			respondError(c, errUnavailable)
			return
		}
	}
	if requestedWidth > 0 && requestedWidth < media.Width {
		original, _, e := image.Decode(bytes.NewReader(data))
		if e != nil {
			respondError(c, e)
			return
		}
		resized := resizeTripImage(original, requestedWidth, int(math.Max(1, math.Round(float64(media.Height)*float64(requestedWidth)/float64(media.Width)))))
		data, e = encodeTripImage(resized, media.ContentType)
		if e != nil {
			respondError(c, e)
			return
		}
	}
	// Access or publication may have changed during the storage read. Recheck
	// under the same locks immediately before writing any protected bytes.
	viewer = s.viewer(c)
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := s.authorizeTripHero(tx, c, viewer, mediaID); err != nil {
			return err
		}
		c.Header("Cache-Control", "private, no-store, max-age=0")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Content-Disposition", "inline")
		c.Data(200, media.ContentType, data)
		return nil
	})
	if err != nil {
		respondError(c, err)
	}
}
func sanitizeTripImage(raw []byte) ([]byte, string, int, int, error) {
	if len(raw) == 0 || len(raw) > maxTripImageBytes {
		return nil, "", 0, 0, bad("Hero images must be JPEG or PNG and no larger than 8 MB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || !oneOf(format, "jpeg", "png") || config.Width <= 0 || config.Height <= 0 || config.Width > maxTripImagePixels/config.Height {
		return nil, "", 0, 0, bad("Use a valid JPEG or PNG image of no more than 20 megapixels")
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", 0, 0, bad("The uploaded image could not be decoded")
	}
	if format == "jpeg" {
		decoded = orientTripImage(decoded, jpegExifOrientation(raw))
	}
	width, height := decoded.Bounds().Dx(), decoded.Bounds().Dy()
	if width > maxTripImageSide || height > maxTripImageSide {
		scale := math.Min(float64(maxTripImageSide)/float64(width), float64(maxTripImageSide)/float64(height))
		width = int(math.Max(1, math.Round(float64(width)*scale)))
		height = int(math.Max(1, math.Round(float64(height)*scale)))
		decoded = resizeTripImage(decoded, width, height)
	}
	contentType := "image/" + format
	if format == "jpeg" {
		contentType = "image/jpeg"
	}
	encoded, err := encodeTripImage(decoded, contentType)
	if err != nil {
		return nil, "", 0, 0, bad("Unable to process this image")
	}
	if len(encoded) > maxTripImageBytes {
		return nil, "", 0, 0, bad("The processed image exceeds 8 MB; choose a smaller image")
	}
	return encoded, contentType, width, height, nil
}
func encodeTripImage(img image.Image, contentType string) ([]byte, error) {
	var out bytes.Buffer
	var err error
	if contentType == "image/png" {
		err = png.Encode(&out, img)
	} else {
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: 88})
	}
	return out.Bytes(), err
}

// Re-encoding pixels discards EXIF/GPS, comments, embedded profiles and other
// original metadata. JPEG orientation is applied before those tags disappear.
func jpegExifOrientation(raw []byte) int {
	if len(raw) < 4 || raw[0] != 0xff || raw[1] != 0xd8 {
		return 1
	}
	for offset := 2; offset+4 <= len(raw); {
		if raw[offset] != 0xff {
			return 1
		}
		marker := raw[offset+1]
		offset += 2
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		if marker == 0xff {
			offset--
			continue
		}
		if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			continue
		}
		length := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
		if length < 2 || offset+length > len(raw) {
			return 1
		}
		segment := raw[offset+2 : offset+length]
		offset += length
		if marker != 0xe1 || len(segment) < 14 || string(segment[:6]) != "Exif\x00\x00" {
			continue
		}
		tiff := segment[6:]
		var order binary.ByteOrder
		switch string(tiff[:2]) {
		case "II":
			order = binary.LittleEndian
		case "MM":
			order = binary.BigEndian
		default:
			return 1
		}
		if order.Uint16(tiff[2:4]) != 42 {
			return 1
		}
		start := int64(order.Uint32(tiff[4:8]))
		if start < 8 || start+2 > int64(len(tiff)) {
			return 1
		}
		entries := int64(order.Uint16(tiff[start : start+2]))
		for i := int64(0); i < entries; i++ {
			p := start + 2 + i*12
			if p+12 > int64(len(tiff)) {
				return 1
			}
			entry := tiff[p : p+12]
			if order.Uint16(entry[:2]) == 0x0112 && order.Uint16(entry[2:4]) == 3 && order.Uint32(entry[4:8]) == 1 {
				orientation := int(order.Uint16(entry[8:10]))
				if orientation >= 1 && orientation <= 8 {
					return orientation
				}
				return 1
			}
		}
	}
	return 1
}
func orientTripImage(src image.Image, orientation int) image.Image {
	if orientation <= 1 || orientation > 8 {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := x, y
			switch orientation {
			case 2:
				dx = w - 1 - x
			case 3:
				dx = w - 1 - x
				dy = h - 1 - y
			case 4:
				dy = h - 1 - y
			case 5:
				dx = y
				dy = x
			case 6:
				dx = h - 1 - y
				dy = x
			case 7:
				dx = h - 1 - y
				dy = w - 1 - x
			case 8:
				dx = y
				dy = w - 1 - x
			}
			dst.Set(dx, dy, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}
func resizeTripImage(src image.Image, width, height int) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	sx, sy := float64(b.Dx())/float64(width), float64(b.Dy())/float64(height)
	for y := 0; y < height; y++ {
		fy := math.Max(0, (float64(y)+.5)*sy-.5)
		y0 := int(fy)
		y1 := min(y0+1, b.Dy()-1)
		wy := fy - float64(y0)
		for x := 0; x < width; x++ {
			fx := math.Max(0, (float64(x)+.5)*sx-.5)
			x0 := int(fx)
			x1 := min(x0+1, b.Dx()-1)
			wx := fx - float64(x0)
			r0, g0, b0, a0 := src.At(b.Min.X+x0, b.Min.Y+y0).RGBA()
			r1, g1, b1, a1 := src.At(b.Min.X+x1, b.Min.Y+y0).RGBA()
			r2, g2, b2, a2 := src.At(b.Min.X+x0, b.Min.Y+y1).RGBA()
			r3, g3, b3, a3 := src.At(b.Min.X+x1, b.Min.Y+y1).RGBA()
			blend := func(a, b, c, d uint32) uint8 {
				return uint8(math.Round(((float64(a)*(1-wx)+float64(b)*wx)*(1-wy) + (float64(c)*(1-wx)+float64(d)*wx)*wy) / 257))
			}
			dst.SetRGBA(x, y, color.RGBA{R: blend(r0, r1, r2, r3), G: blend(g0, g1, g2, g3), B: blend(b0, b1, b2, b3), A: blend(a0, a1, a2, a3)})
		}
	}
	return dst
}
