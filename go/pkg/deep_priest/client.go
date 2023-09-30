package deep_priest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type deepPriest struct{}

type DeepPriest interface {
	PetitionTheFount(*Question) (*Answer, error)
}

type Question struct {
	Question string `json:"question"`
}

type Answer struct {
	Answer string `json:"answer"`
}

const (
	baseUrl = "http://localhost:8081"
)

func SummonDeepPriest() DeepPriest {
	return &deepPriest{}
}

func (d *deepPriest) PetitionTheFount(question *Question) (*Answer, error) {
	jsonBody, err := json.Marshal(question)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/consult", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var answer Answer
	err = json.Unmarshal(body, &answer)
	if err != nil {
		return nil, err
	}

	return &answer, nil
}
