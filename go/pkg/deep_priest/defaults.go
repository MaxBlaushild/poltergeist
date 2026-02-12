package deep_priest

const (
	DefaultImageModel   = "gpt-image-1"
	DefaultImageSize    = "1024x1024"
	DefaultImageQuality = "auto"
	DefaultImageUser    = "poltergeist"
	DefaultImageN       = 1
)

func ApplyGenerateImageDefaults(req *GenerateImageRequest) {
	if req.Model == "" {
		req.Model = DefaultImageModel
	}
	if req.Size == "" {
		req.Size = DefaultImageSize
	}
	if req.Quality == "" {
		req.Quality = DefaultImageQuality
	}
	if req.User == "" {
		req.User = DefaultImageUser
	}
	if req.N == 0 {
		req.N = DefaultImageN
	}
	req.Size = normalizeImageSize(req.Size)
	req.Quality = normalizeImageQuality(req.Quality)
	// response_format is rejected by the current image backend; keep it unset.
	if req.ResponseFormat != "" {
		req.ResponseFormat = ""
	}
}

func ApplyEditImageDefaults(req *EditImageRequest) {
	if req.Model == "" {
		req.Model = DefaultImageModel
	}
	if req.Size == "" {
		req.Size = DefaultImageSize
	}
	if req.Quality == "" {
		req.Quality = DefaultImageQuality
	}
	if req.User == "" {
		req.User = DefaultImageUser
	}
	if req.N == 0 {
		req.N = DefaultImageN
	}
	req.Size = normalizeImageSize(req.Size)
	req.Quality = normalizeImageQuality(req.Quality)
	if req.ResponseFormat != "" {
		req.ResponseFormat = ""
	}
}

func normalizeImageSize(size string) string {
	switch size {
	case "1024x1024", "1024x1536", "1536x1024", "auto":
		return size
	default:
		return DefaultImageSize
	}
}

func normalizeImageQuality(quality string) string {
	switch quality {
	case "low", "medium", "high", "auto":
		return quality
	default:
		return DefaultImageQuality
	}
}
