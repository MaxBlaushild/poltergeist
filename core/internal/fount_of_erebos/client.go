package fount_of_erebos

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client interface {
	ConsultTheDeep(question *DesperateQuery) (*DarkKnowing, error)
}

type client struct {
	baseUrl string
}

type DesperateQuery struct {
	Question string `json:"question"`
}

type DarkKnowing struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func NewClient(baseUrl string) Client {
	return &client{baseUrl: baseUrl}
}

func (c *client) ConsultTheDeep(question *DesperateQuery) (*DarkKnowing, error) {
	resp, err := c.post("/consult", question)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var k DarkKnowing
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&k); err != nil {
		return nil, err
	}

	k.Question = question.Question

	return &k, nil
}

func (c *client) post(endpoint string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseUrl+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	return resp, nil
}
