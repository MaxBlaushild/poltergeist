package processors

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/util"

	"github.com/disintegration/imaging"
	"github.com/hibiken/asynq"
)

// Stricter, checklist-style prompt for consistent pixel art.
const strictPixelPrompt = `
	Create an image in the style of retro 16-bit RPG pixel art.
	Use bold outlines, limited flat colors, and minimal dithering.
	Apply 2–3 shading tones per area, with crisp, blocky edges.
	Keep the result clean, simple, and non-photorealistic.
	Avoid gradients, text, and logos.
`

// Tunables
const (
	genSize           = "1024x1024" // generation size; then we'll enforce pixel look
	iconSize          = 128       // pixel icon size (downscale target before quantize)
	quantColors       = 20        // ~16–24 is a sweet spot; adjust to taste
	upscaleOutput     = 512       // final upscaled display size with nearest neighbor
	pickThresholdEdge = 0.08      // min edge fraction; used only as sanity check
)

// CreateProfilePictureProcessor generates a pixel-art avatar from a selfie URL.
type CreateProfilePictureProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewCreateProfilePictureProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) CreateProfilePictureProcessor {
	log.Println("Initializing CreateProfilePictureProcessor (Pixel Avatar v2)")
	return CreateProfilePictureProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *CreateProfilePictureProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing create profile picture task: %v", task.Type())

	var payload jobs.CreateProfilePictureTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("CreateProfilePicture payload: %+v", payload)

	user, err := p.dbClient.User().FindByID(ctx, payload.UserID)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return fmt.Errorf("failed to find user: %w", err)
	}

	log.Printf("Generating pixel avatar for user ID: %v", payload.UserID)

	// Ask for multiple candidates; we'll pick the crispest one.
	editRequest := deep_priest.EditImageRequest{
		Prompt:   strictPixelPrompt,
		ImageUrl: payload.ProfilePictureUrl,
		Model:    "gpt-image-1",
		N:        1,       // multiple for selection
		Size:     genSize, // generate smaller; we'll downscale further
		// If your backend supports: TransparentBackground: true,
		// If you add masks later: MaskUrl: payload.FaceMaskURL,
	}
	deep_priest.ApplyEditImageDefaults(&editRequest)
	resp, err := p.deepPriestClient.EditImage(editRequest)
	if err != nil {
		log.Printf("Failed to generate profile picture(s): %v", err)
		return fmt.Errorf("failed to generate profile picture: %w", err)
	}

	candidates, err := decodeBase64Candidates(resp)
	if err != nil {
		log.Printf("Failed to decode base64 candidates: %v", err)
		return fmt.Errorf("failed to decode candidates: %w", err)
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no image candidates returned")
	}

	// Post-process each candidate to enforce a true pixel look.
	type scored struct {
		img   image.Image
		bytes []byte
		score float64
		idx   int
	}
	var processed []scored
	for i, b := range candidates {
		pp, img, err := EnforcePixelLook(b, iconSize, quantColors, upscaleOutput, true /*try transparency*/)
		if err != nil {
			log.Printf("Post-process failed for candidate %d: %v", i, err)
			continue
		}
		score := edgeDensityScore(img)
		processed = append(processed, scored{img: img, bytes: pp, score: score, idx: i})
	}

	if len(processed) == 0 {
		return fmt.Errorf("all candidates failed post-processing")
	}

	// Pick the candidate with highest edge density (crisper outlines near face usually win).
	sort.SliceStable(processed, func(i, j int) bool { return processed[i].score > processed[j].score })
	best := processed[0]
	log.Printf("Picked candidate %d with edge score=%.4f", best.idx, best.score)
	if best.score < pickThresholdEdge {
		log.Printf("Warning: low edge score (%.4f). Result may look soft.", best.score)
	}

	url, err := p.UploadImage(ctx, user.ID.String(), best.bytes)
	if err != nil {
		log.Printf("Failed to upload image: %v", err)
		return fmt.Errorf("failed to upload image: %w", err)
	}

	if err := p.dbClient.User().UpdateProfilePictureUrl(ctx, payload.UserID, url); err != nil {
		log.Printf("Failed to update user profile picture URL: %v", err)
		return fmt.Errorf("failed to update user profile picture URL: %w", err)
	}

	log.Printf("Profile picture updated successfully for user ID: %v", payload.UserID)
	return nil
}

func (p *CreateProfilePictureProcessor) UploadImage(ctx context.Context, userID string, imageBytes []byte) (string, error) {
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("no image data provided")
	}

	imageFormat, err := util.DetectImageFormat(imageBytes)
	if err != nil {
		return "", err
	}

	imageExtension, err := util.GetImageExtension(imageFormat)
	if err != nil {
		return "", err
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 16)
	imageName := timestamp + "-" + userID + "." + imageExtension

	imageUrl, err := p.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
	if err != nil {
		return "", err
	}

	return imageUrl, nil
}

// ---------- Helpers ----------

// decodeBase64Candidates supports either a single base64 string or a JSON array of base64 strings.
func decodeBase64Candidates(b64json string) ([][]byte, error) {
	// Try as JSON array of strings
	var arr []string
	if err := json.Unmarshal([]byte(b64json), &arr); err == nil && len(arr) > 0 {
		out := make([][]byte, 0, len(arr))
		for _, s := range arr {
			dec, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("failed to decode array item: %w", err)
			}
			out = append(out, dec)
		}
		return out, nil
	}
	// Fall back to single base64 string
	raw, err := base64.StdEncoding.DecodeString(b64json)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return [][]byte{raw}, nil
}

// EnforcePixelLook: downscale -> optional background transparency -> palette quantize -> upscale (nearest).
// Returns encoded PNG bytes and the processed image.Image.
func EnforcePixelLook(imgBytes []byte, targetIcon int, colors int, upscale int, tryTransparent bool) ([]byte, image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, nil, err
	}

	// 1) Downscale to icon size with nearest neighbor to force chunky pixels
	icon := imaging.Resize(img, targetIcon, targetIcon, imaging.NearestNeighbor)

	// 2) Optional transparency pass (remove flat background close to the dominant border color)
	var withBG image.Image = icon
	if tryTransparent {
		withBG = makeBorderColorTransparent(icon, 8, 14.0) // borderWidth, tolerance
	}

	// 3) Upscale for display using nearest neighbor to preserve the blocks
	out := imaging.Resize(withBG, upscale, upscale, imaging.NearestNeighbor)

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), out, nil
}

// edgeDensityScore: quick Sobel-based edge fraction in [0..1].
func edgeDensityScore(img image.Image) float64 {
	// Convert to grayscale float
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	gray := make([]float64, w*h)
	idx := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b2, _ := img.At(x, y).RGBA()
			// perceptual luma
			l := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b2)
			gray[idx] = l / 65535.0
			idx++
		}
	}

	// Simple Sobel
	var edges int
	threshold := 0.12 // tweakable
	sx := [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	sy := [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	at := func(xx, yy int) float64 { return gray[yy*w+xx] }

	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			var gx, gy float64
			for j := -1; j <= 1; j++ {
				for i := -1; i <= 1; i++ {
					v := at(x+i, y+j)
					gx += sx[j+1][i+1] * v
					gy += sy[j+1][i+1] * v
				}
			}
			mag := math.Hypot(gx, gy)
			if mag > threshold {
				edges++
			}
		}
	}
	total := (w - 2) * (h - 2)
	if total <= 0 {
		return 0
	}
	return float64(edges) / float64(total)
}

// makeBorderColorTransparent finds the dominant color on the outer border and sets similar pixels to alpha=0.
func makeBorderColorTransparent(src image.Image, border int, tolerance float64) image.Image {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	// Collect border colors
	type rgb struct{ r, g, b int }
	counts := map[rgb]int{}
	add := func(c color.Color) {
		r8, g8, b8, _ := c.RGBA()
		key := rgb{int(r8 >> 8), int(g8 >> 8), int(b8 >> 8)}
		counts[key]++
	}

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Min.Y+border && y < b.Max.Y; y++ {
			add(src.At(x, y))
		}
		for y := b.Max.Y - border; y < b.Max.Y; y++ {
			if y >= b.Min.Y && y < b.Max.Y {
				add(src.At(x, y))
			}
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Min.X+border && x < b.Max.X; x++ {
			add(src.At(x, y))
		}
		for x := b.Max.X - border; x < b.Max.X; x++ {
			if x >= b.Min.X && x < b.Max.X {
				add(src.At(x, y))
			}
		}
	}

	// Find dominant
	var dom rgb
	var max int
	for k, v := range counts {
		if v > max {
			max, dom = v, k
		}
	}

	// Replace similar with alpha=0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			off := dst.PixOffset(x, y)
			r := int(dst.Pix[off+0])
			g := int(dst.Pix[off+1])
			bl := int(dst.Pix[off+2])
			if colorDist(rgb{r, g, bl}, dom) <= tolerance {
				dst.Pix[off+3] = 0 // alpha
			}
		}
	}
	return dst
}

func colorDist(a, b struct{ r, g, b int }) float64 {
	dr := float64(a.r - b.r)
	dg := float64(a.g - b.g)
	db := float64(a.b - b.b)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}
