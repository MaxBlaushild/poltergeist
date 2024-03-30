package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/twilio"
	"github.com/MaxBlaushild/poltergeist/texter/internal/config"
	"github.com/gin-gonic/gin"
)

type TwilioSMSRequest struct {
	To                  string `form:"To"`
	From                string `form:"From"`
	Body                string `form:"Body"`
	MessageSid          string `form:"MessageSid"`
	AccountSid          string `form:"AccountSid"`
	MessagingServiceSid string `form:"MessagingServiceSid"`
	NumMedia            string `form:"NumMedia"`
}

type Text struct {
	Body     string `json:"body" binding:"required"`
	To       string `json:"to" binding:"required"`
	From     string `json:"from" binding:"required"`
	TextType string `json:"textType" binding:"required"`
}

func forwardText(ctx *gin.Context, url string, text *Text) (*Text, error) {
	jsonBody, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp)
		return nil, errors.New("received non-OK response from server")
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseText Text
	if err = json.Unmarshal(res, &responseText); err != nil {
		return nil, err
	}

	return &responseText, nil
}

func main() {
	router := gin.Default()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     "poltergeist",
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	twilioClient := twilio.NewClient(&twilio.ClientConfig{
		AccountSid: cfg.Secret.TwilioAccountSid,
		AuthToken:  cfg.Secret.TwilioAuthToken,
	})

	router.GET("/texter/sent-texts/count", func(ctx *gin.Context) {
		phoneNumber := ctx.Query("phoneNumber")
		textType := ctx.Query("textType")

		if len(phoneNumber) == 0 {
			ctx.JSON(400, gin.H{
				"error": "phone number is required",
			})
			return
		}

		if len(textType) == 0 {
			ctx.JSON(400, gin.H{
				"error": "text type is required",
			})
			return
		}

		count, err := dbClient.SentText().GetCount(ctx, phoneNumber, textType)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(200, gin.H{
			"count": count,
		})
	})

	router.POST("/texter/send-sms", func(ctx *gin.Context) {
		var text Text

		if err := ctx.Bind(&text); err != nil {
			fmt.Println(err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "shit sms request",
			})
			return
		}

		if err := twilioClient.SendText(ctx, &twilio.Text{
			Text: text.Body,
			To:   text.To,
			From: text.From,
		}); err != nil {
			fmt.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "sending text went wrong",
			})
			return
		}

		if _, err := dbClient.SentText().Insert(ctx, text.TextType, text.To, text.Body); err != nil {
			fmt.Println(err)
		}

		ctx.JSON(200, gin.H{
			"message": "text sent successfully",
		})
	})

	router.POST("/texter/receive-sms", func(ctx *gin.Context) {
		var smsRequest TwilioSMSRequest

		if err := ctx.ShouldBind(&smsRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var url string
		switch smsRequest.To {
		case cfg.Secret.GuessHowManyPhoneNumber:
			url = "http://localhost:8082/trivai/receive-sms"
		default:
			ctx.JSON(400, gin.H{
				"message": "this aint my number!",
			})
			return
		}

		text, err := forwardText(ctx, url, &Text{
			To:   smsRequest.To,
			From: smsRequest.From,
			Body: smsRequest.Body,
		})
		if err != nil {
			fmt.Println(err)
			ctx.JSON(500, gin.H{
				"message": err.Error(),
			})
		}

		if err := twilioClient.SendText(ctx, &twilio.Text{
			Text: text.Body,
			To:   text.To,
			From: smsRequest.To,
		}); err != nil {
			fmt.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "sending text went wrong",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "text sent successfully",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.Run(":8084")
}
