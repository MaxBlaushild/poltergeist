package open_ai

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	openai "github.com/sashabaranov/go-openai"
)

type client struct {
	ai *openai.Client
}

type ClientConfig struct {
	ApiKey string
}

func NewClient(config ClientConfig) OpenAiClient {
	log.Println("Initializing OpenAI client")
	ai := openai.NewClient(config.ApiKey)

	return &client{
		ai: ai,
	}
}

type neverOpaque struct{ image.Image }

func (neverOpaque) Opaque() bool { return false }

// convertToPNG converts image data to PNG format with RGBA color space
func convertToPNG(imageData []byte) ([]byte, error) {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Check if it's already RGBA
	if rgbaImg, ok := img.(*image.RGBA); ok {
		// Encode directly
		var pngBuffer bytes.Buffer
		err = png.Encode(&pngBuffer, neverOpaque{rgbaImg})
		if err != nil {
			return nil, fmt.Errorf("failed to encode RGBA image as PNG: %w", err)
		}
		return pngBuffer.Bytes(), nil
	}

	// Try using NRGBA format (non-alpha-premultiplied RGBA)
	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)

	// Copy image data to NRGBA format pixel by pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			// Convert to NRGBA
			r, g, b, a := originalColor.RGBA()
			nrgbaImg.Set(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	// Encode as PNG
	var pngBuffer bytes.Buffer
	err = png.Encode(&pngBuffer, neverOpaque{nrgbaImg})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	log.Printf("Successfully converted image to PNG with alpha channel")
	return pngBuffer.Bytes(), nil
}

func (c *client) GetAnswer(ctx context.Context, q string) (string, error) {
	log.Printf("Getting answer for question: %s", q)
	resp, err := c.ai.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Temperature: 0.1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: q,
				},
			},
		},
	)

	if err != nil {
		log.Printf("Error getting answer: %v", err)
		return "", err
	}

	log.Printf("Successfully got answer: %s", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func (c *client) GenerateImage(ctx context.Context, request deep_priest.GenerateImageRequest) (string, error) {
	log.Printf("Generating image with prompt: %s", request.Prompt)
	resp, err := c.ai.CreateImage(
		ctx,
		openai.ImageRequest{
			Prompt:         request.Prompt,
			N:              request.N,
			Size:           request.Size,
			Style:          request.Style,
			User:           request.User,
			Quality:        request.Quality,
			ResponseFormat: request.ResponseFormat,
			Model:          request.Model,
		},
	)

	if err != nil {
		log.Printf("Error generating image: %v", err)
		return "", err
	}

	log.Printf("Successfully generated image")
	return resp.Data[0].B64JSON, nil
}

func (c *client) EditImage(ctx context.Context, request deep_priest.EditImageRequest) (string, error) {
	log.Printf("Editing image with prompt: %q", request.Prompt)
	log.Printf("Image URL: %s", request.ImageUrl)

	// 1) Download the source image
	resp, err := http.Get(request.ImageUrl)
	if err != nil {
		log.Printf("Error downloading image: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error downloading image: status code %d", resp.StatusCode)
		return "", fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading image data: %v", err)
		return "", err
	}

	c.ai.CreateFile(ctx, openai.FileRequest{
		Purpose: "vision",
	})

	// 2) Convert the image to PNG (what the Images Edit endpoint expects)
	pngData, err := convertToPNG(imageData)
	if err != nil {
		log.Printf("Error converting image to PNG: %v", err)
		return "", err
	}

	// 3) Decode (the PNG) just to get bounds for mask sizing
	imgForBounds, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Printf("Error decoding PNG for bounds: %v", err)
		return "", err
	}
	bounds := imgForBounds.Bounds()

	// 4) Build a fully transparent mask of the same size.
	//    In OpenAI image edits: transparent (A=0) == editable, opaque (A=255) == preserved.
	mask := image.NewNRGBA(bounds)
	// NRGBA is zero-initialized, so A=0 everywhere already. If you want to be explicit:
	// for i := 3; i < len(mask.Pix); i += 4 { mask.Pix[i] = 0 }

	var maskBuf bytes.Buffer
	if err := png.Encode(&maskBuf, mask); err != nil {
		log.Printf("Error encoding mask PNG: %v", err)
		return "", err
	}

	// 5) Wrap both files for multipart upload
	imageReader := bytes.NewReader(pngData)
	wrappedImage := openai.WrapReader(imageReader, "image.png", "image/png")

	maskReader := bytes.NewReader(maskBuf.Bytes())
	wrappedMask := openai.WrapReader(maskReader, "mask.png", "image/png")

	log.Printf("Mask data length: %d", len(maskBuf.Bytes()))
	log.Printf("image url: %s", request.ImageUrl)
	log.Printf("Image edit request details:\n"+
		"  Prompt: %s\n"+
		"  Model: %s\n"+
		"  N: %d\n"+
		"  Quality: %s\n"+
		"  Size: %s\n"+
		"  Style: %s\n"+
		"  ResponseFormat: %s\n"+
		"  User: %s\n"+
		"  Image URL: %s",
		request.Prompt,
		request.Model,
		request.N,
		request.Quality,
		request.Size,
		request.Style,
		request.ResponseFormat,
		request.User,
		request.ImageUrl)

	// 6) Call the image edit endpoint WITH the transparent mask
	newImage, err := c.ai.CreateEditImage(
		ctx,
		openai.ImageEditRequest{
			Prompt:         request.Prompt,
			Image:          wrappedImage,
			Mask:           wrappedMask, // <-- the important part
			Model:          request.Model,
			N:              request.N,
			Quality:        request.Quality,
			Size:           request.Size,
			ResponseFormat: request.ResponseFormat,
			User:           request.User,
		},
	)
	if err != nil {
		log.Printf("Error editing image: %v", err)
		return "", err
	}

	// Note: this is base64 PNG data, not a URL.
	return newImage.Data[0].B64JSON, nil
}

func (c *client) GetAnswerWithImage(ctx context.Context, q string, imageUrl string) (string, error) {
	log.Printf("Getting answer for question with image. Question: %s, Image URL: %s", q, imageUrl)
	resp, err := c.ai.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Temperature: 0.1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: q,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL:    imageUrl,
								Detail: openai.ImageURLDetailAuto,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		log.Printf("Error getting answer with image: %v", err)
		return "", err
	}

	log.Printf("Successfully got answer with image: %s", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

// func (c *client) GenerateImageWithImage(ctx context.Context, request deep_priest.EditImageRequest) (string, error) {
// 	imgResp, err := http.Get(request.ImageUrl)
// 	if err != nil {
// 		log.Printf("Error getting image: %v", err)
// 		return "", err
// 	}
// 	defer imgResp.Body.Close()

// 	imgData, err := io.ReadAll(imgResp.Body)
// 	if err != nil {
// 		log.Printf("Error reading image data: %v", err)
// 		return "", err
// 	}

// 	// Determine the MIME type from the URL extension
// 	ext := filepath.Ext(request.ImageUrl)
// 	mimeType := mime.TypeByExtension(ext)
// 	if mimeType == "" {
// 		// Default to JPEG if we can't determine the type
// 		mimeType = "image/jpeg"
// 	}

// 	// Encode image in base64 for GPT-4o multimodal input
// 	imgBase64 := base64.StdEncoding.EncodeToString(imgData)
// 	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, imgBase64)

// 	// 2️⃣ Ask GPT-4o to craft a perfect pixel-art prompt
// 	chatReq := openai.ChatCompletionRequest{
// 		Model: openai.GPT4o,
// 		Messages: []openai.ChatCompletionMessage{
// 			{
// 				Role:    openai.ChatMessageRoleSystem,
// 				Content: "You are an art director who writes perfect DALL-E prompts.",
// 			},
// 			{
// 				Role:    openai.ChatMessageRoleUser,
// 				Content: request.Prompt,
// 			},
// 			{
// 				Role: openai.ChatMessageRoleUser,
// 				MultiContent: []openai.ChatMessagePart{
// 					{
// 						Type: "image_url",
// 						ImageURL: &openai.ChatMessageImageURL{
// 							URL: dataURL,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	chatResp, err := c.ai.CreateChatCompletion(ctx, chatReq)
// 	if err != nil {
// 		log.Printf("Error generating chat completion: %v", err)
// 		return "", err
// 	}

// 	prompt := chatResp.Choices[0].Message.Content
// 	fmt.Println("🎨 Generated pixel-art prompt:\n", prompt)

// 	imgGenReq := openai.ImageRequest{
// 		Prompt:         prompt,
// 		Model:          request.Model,
// 		N:              request.N,
// 		Quality:        request.Quality,
// 		Size:           request.Size,
// 		ResponseFormat: request.ResponseFormat,
// 		User:           request.User,
// 	}

// 	genResp, err := c.ai.CreateImage(ctx, imgGenReq)
// 	if err != nil {
// 		log.Printf("Error generating image: %v", err)
// 		return "", err
// 	}

// 	// 4️⃣ Decode and save the generated image
// 	imgDecoded, err := base64.StdEncoding.DecodeString(genResp.Data[0].B64JSON)
// 	if err != nil {
// 		log.Printf("Error decoding image: %v", err)
// 		return "", err
// 	}

// 	if err := os.WriteFile("pixel_portrait.png", imgDecoded, 0644); err != nil {
// 		log.Printf("Error writing image: %v", err)
// 		return "", err
// 	}

// 	fmt.Println("✅ Saved pixelated RPG portrait as pixel_portrait.png")
// 	return genResp.Data[0].B64JSON, nil
// }
