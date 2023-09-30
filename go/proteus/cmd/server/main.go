package main

import (
	"net/http"

	"github.com/MaxBlaushild/proteus/internal/hue"
	"github.com/gin-gonic/gin"
)

func main() {
	// ctx := context.Background()

	// devices, err := device_discovery.DiscoverDevices(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// frameClient, err := frame.NewFrameClient(devices[0])
	// if err != nil {
	// 	panic(err)
	// }

	hueClient, err := hue.NewClient()
	if err != nil {
		panic(err)
	}

	// errors := make(chan error)
	// done := make(chan bool)

	router := gin.Default()

	router.GET("/lights/on", func(ctx *gin.Context) {
		if err := hueClient.TurnOnLights(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "something went wrong",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "lights are on",
		})
		return
	})

	router.GET("/lights/off", func(ctx *gin.Context) {
		if err := hueClient.TurnOffLights(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "something went wrong",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "lights are off",
		})
		return
	})

	// router.GET("/artMode/on", func(ctx *gin.Context) {
	// 	if err := frameClient.ToggleArtMode("on"); err != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 			"message": "something went wrong",
	// 		})
	// 		return
	// 	}

	// 	ctx.JSON(200, gin.H{
	// 		"message": "art mode is on",
	// 	})
	// 	return
	// })

	// router.GET("/artMode/off", func(ctx *gin.Context) {
	// 	if err := frameClient.ToggleArtMode("off"); err != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 			"message": "something went wrong",
	// 		})
	// 		return
	// 	}

	// 	ctx.JSON(200, gin.H{
	// 		"message": "art mode is off",
	// 	})
	// 	return
	// })

	router.Run(":8085")

	// go frameClient.Connect(ctx, done, errors)

	// for {
	// 	select {
	// 	case <-done:
	// 		fmt.Println("Frame TV ready for requests")
	// 		// frameClient.RequestSendImage("./test.png")

	// 	case err := <-errors:
	// 		panic(err)
	// 	case <-ctx.Done():
	// 		panic(ctx.Err())
	// 	}
	// }
}
