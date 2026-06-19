package server

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxPhotosPerSubmission = 6
	maxPhotoBytes          = 5 << 20 // 5 MB after decode (phone-resized JPEGs are far smaller)
)

// decodeDataURL parses a "data:image/jpeg;base64,..." string into its content
// type and raw bytes.
func decodeDataURL(s string) (string, []byte, error) {
	if !strings.HasPrefix(s, "data:") {
		return "", nil, errors.New("not a data url")
	}
	comma := strings.IndexByte(s, ',')
	if comma < 0 {
		return "", nil, errors.New("malformed data url")
	}
	meta := s[len("data:"):comma] // e.g. image/jpeg;base64
	ct := meta
	if i := strings.IndexByte(meta, ';'); i >= 0 {
		ct = meta[:i]
	}
	if !strings.HasPrefix(ct, "image/") {
		return "", nil, errors.New("not an image")
	}
	data, err := base64.StdEncoding.DecodeString(s[comma+1:])
	if err != nil {
		return "", nil, err
	}
	return ct, data, nil
}

// GET /photos/:id — serve a submission photo's bytes. The id is an unguessable
// UUID; these are mission-proof snapshots, not secret content, so no auth gate
// (which also lets the browser load them via plain <img src>).
func (s *server) getPhoto(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo id"})
		return
	}
	photo, err := s.dbClient.Vampire().GetPhoto(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if photo == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.Header("Cache-Control", "private, max-age=86400")
	ctx.Data(http.StatusOK, photo.ContentType, photo.Data)
}

// savePhotos applies a submission's photo edits: optional clear, then append the
// provided data-URLs up to the per-submission cap.
func (s *server) savePhotos(ctx *gin.Context, submissionID uuid.UUID, dataURLs []string, clear bool) error {
	v := s.dbClient.Vampire()
	if clear {
		if err := v.DeletePhotosForSubmission(ctx, submissionID); err != nil {
			return err
		}
	}

	// Count what's already there so appends respect the cap.
	refs, err := v.ListPhotoRefs(ctx)
	if err != nil {
		return err
	}
	count := 0
	for _, r := range refs {
		if r.SubmissionID == submissionID {
			count++
		}
	}

	for _, durl := range dataURLs {
		if count >= maxPhotosPerSubmission {
			break
		}
		ct, data, err := decodeDataURL(durl)
		if err != nil || len(data) == 0 || len(data) > maxPhotoBytes {
			continue // skip bad/oversized images rather than failing the whole submit
		}
		if _, err := v.AddSubmissionPhoto(ctx, submissionID, ct, data); err != nil {
			return err
		}
		count++
	}
	return nil
}

// photoIDsBySubmission builds submissionID -> [photo url-id] for the views.
func (s *server) photoIDsBySubmission(ctx *gin.Context) (map[string][]string, error) {
	refs, err := s.dbClient.Vampire().ListPhotoRefs(ctx)
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, r := range refs {
		sid := r.SubmissionID.String()
		out[sid] = append(out[sid], r.ID.String())
	}
	return out, nil
}
