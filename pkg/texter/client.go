package texter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Text struct {
	Body string `json:"body" binding:"required"`
	To   string `json:"to" binding:"required"`
	From string `json:"from" binding:"required"`
}

type texterClient struct{}

type TexterClient interface {
	Text(*Text) error
}

const (
	baseUrl = "http://localhost:8084"
)

func NewTexterClient() TexterClient {
	return &texterClient{}
}

func (d *texterClient) Text(text *Text) error {
	jsonBody, err := json.Marshal(text)
	if err != nil {
		return err
	}

	resp, err := http.Post(baseUrl+"/texter/send-sms", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
