package dungeonmaster

import (
	"context"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
)

func (c *client) UploadImage(ctx context.Context, base64Image string) (string, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return "", err
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
	imageName := timestamp + "-" + uuid.New().String() + "." + imageExtension

	imageUrl, err := c.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
	if err != nil {
		return "", err
	}

	return imageUrl, nil
}
