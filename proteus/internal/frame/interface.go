package frame

import "context"

type FrameClient interface {
	Connect(ctx context.Context, done chan bool, errors chan error)
	ToggleArtMode(status string) error
	RequestSendImage(pathToImage string) error
}
