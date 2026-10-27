package travelcalendar

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TripLodgingPhoto struct {
	MediaID uuid.UUID `json:"mediaId"`
	AltText string    `json:"altText"`
	URL     string    `json:"url"`
	Width   int       `json:"width"`
	Height  int       `json:"height"`
}

func lodgingPhotoURL(stopID uuid.UUID, entryID string, mediaID uuid.UUID) string {
	return "/travel-angels/calendar/stops/" + stopID.String() + "/lodging/" + entryID + "/photos/" + mediaID.String()
}

func (s *Service) uploadLodgingPhoto(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	entryID, err := uuid.Parse(c.Param("entryId"))
	if err != nil || entryID == uuid.Nil {
		respondError(c, errUnavailable)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTripImageBytes+(64<<10))
	if err = c.Request.ParseMultipartForm(maxTripImageBytes); err != nil {
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
		respondError(c, bad("Lodging photos must be no larger than 8 MB"))
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
	altText := strings.TrimSpace(c.PostForm("altText"))
	if altText == "" || utf8.RuneCountInString(altText) > 500 {
		respondError(c, bad("Describe this lodging photo in up to 500 characters"))
		return
	}
	var revision *int
	if value := c.PostForm("revision"); value != "" {
		parsed, e := strconv.Atoi(value)
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
		var entry *TripLodging
		for i := range stop.Draft.Lodging {
			if stop.Draft.Lodging[i].ID == entryID.String() {
				entry = &stop.Draft.Lodging[i]
				break
			}
		}
		if entry == nil {
			return bad("Save the lodging entry before adding photos")
		}
		if len(entry.Photos) >= 6 {
			return bad("Each lodging supports up to 6 uploaded photos")
		}
		media := TripMedia{ID: uuid.New(), StopID: stop.ID, LodgingID: &entryID, UploadedBy: user.ID, ContentType: contentType, Width: width, Height: height, Data: sanitized, CreatedAt: time.Now().UTC()}
		if s.tripImageStorage != nil && s.tripImageBucket != "" {
			extension := ".jpg"
			if contentType == "image/png" {
				extension = ".png"
			}
			media.StorageBucket = s.tripImageBucket
			media.StorageKey = "trip-lodging/" + stop.ID.String() + "/" + entry.ID + "/" + media.ID.String() + extension
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
		entry.Photos = append(entry.Photos, TripLodgingPhoto{MediaID: media.ID, AltText: altText, URL: lodgingPhotoURL(stop.ID, entry.ID, media.ID), Width: width, Height: height})
		if err = s.saveTripChange(tx, stop, user.ID, "lodging_photo_uploaded", gin.H{"entryId": entry.ID, "mediaId": media.ID}); err != nil {
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

func (s *Service) authorizeLodgingPhoto(tx *gorm.DB, c *gin.Context, viewer Viewer, mediaID uuid.UUID) error {
	_, entry, err := s.authorizedLodging(tx, c, viewer)
	if err != nil {
		return err
	}
	for _, photo := range entry.Photos {
		if photo.MediaID == mediaID {
			return nil
		}
	}
	return errUnavailable
}

func (s *Service) getLodgingPhoto(c *gin.Context) {
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		respondError(c, errUnavailable)
		return
	}
	viewer := s.viewer(c)
	var media TripMedia
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := s.authorizeLodgingPhoto(tx, c, viewer, mediaID); err != nil {
			return err
		}
		return tx.Where("id = ? AND stop_id = ? AND lodging_id = ?", mediaID, c.Param("id"), c.Param("entryId")).First(&media).Error
	})
	if err != nil {
		respondError(c, err)
		return
	}
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
	}
	if len(data) == 0 || len(data) > maxTripImageBytes {
		respondError(c, errUnavailable)
		return
	}
	viewer = s.viewer(c)
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := s.authorizeLodgingPhoto(tx, c, viewer, mediaID); err != nil {
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
