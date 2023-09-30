package device_discovery

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/huin/goupnp"
)

type Device struct {
	USN string
	URL *url.URL
}

func DiscoverDevices(ctx context.Context) ([]Device, error) {
	devices, err := goupnp.DiscoverDevicesCtx(ctx, "ssdp:all")
	if err != nil {
		return nil, err
	}

	var d []Device
	for _, dev := range devices {
		fmt.Println(dev.Location)
		if strings.Contains(dev.USN, livingRoomFrameTV) {
			d = append(d, Device{
				USN: dev.USN,
				URL: dev.Location,
			})
		}

	}

	return d, nil
}
