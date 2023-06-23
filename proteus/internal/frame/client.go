package frame

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/MaxBlaushild/proteus/internal/device_discovery"
	"github.com/pkg/errors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	method = "ms.channel.emit"
	event  = "art_app_request"
	to     = "host"

	setArtmodeStatus = "set_artmode_status"
	sendImage        = "send_image"

	connectEvent = "ms.channel.connect"
	readyEvent   = "ms.channel.ready"
)

type frameClient struct {
	url  *url.URL
	msgs chan []byte
	ID   string
}

type artModeData struct {
	Value   string `json:"value"`
	Request string `json:"request"`
	ID      string `json:"id"`
}

type sendImageRequest struct {
	Request   string `json:"request"`
	FileType  string `json:"file_type"`
	ImageDate string `json:"image_date"`
	FileSize  int    `json:"file_size"`
	ConnInfo  string `json:"conn_info"`
}

type connectionData struct {
	D2dMode      string `json:"d2d_mode"`
	ConnectionID int    `json:"connection_id"`
	RequestID    string `json:"request_id,omitempty"`
	ID           string `json:"id"`
}

// String d2d_mode = "socket";
//             // this is a random number usually
//             long connection_id = 2705890518L;
//             String request_id;
//             String id = remoteControllerWebSocket.uuid.toString();

type frameTvRequest struct {
	Method string               `json:"method"`
	Params frameTvRequestParams `json:"params"`
}

type frameTvRequestParams struct {
	To    string `json:"to"`
	Event string `json:"event"`
	Data  string `json:"data"`
}

type frameTvResponse struct {
	Data  interface{}
	Event string
}

// map[clients:[map[attributes:map[name:<nil>] connectTime:1.686970300676e+12 deviceName:Smart Device id:f51b57cd-efab-4dee-a911-7181aa884914 isHost:true] map[attributes:map[name:poltergeist] connectTime:1.686971490807e+12 deviceName:poltergeist id:aba699a4-fd35-4973-baf4-d5616af4c895 isHost:false]] id:aba699a4-fd35-4973-baf4-d5616af4c895]
type frameTvConnectResponse struct {
	ID string `json:"id"`
}

func NewFrameClient(device device_discovery.Device) (FrameClient, error) {
	return &frameClient{url: device.URL}, nil
}

func (f *frameClient) Connect(ctx context.Context, done chan bool, errors chan error) {
	c, _, err := websocket.Dial(
		ctx,
		fmt.Sprintf(`ws://%s:8001/api/v2/channels/com.samsung.art-app?name=%s`,
			f.url.Hostname(),
			"poltergeist",
		),
		nil,
	)
	if err != nil {
		errors <- err
	}

	f.msgs = make(chan []byte)

	go f.WaitForResponses(ctx, c, done, errors)

	for {
		select {
		case msg := <-f.msgs:
			fmt.Println(string(msg))
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				errors <- err
			}
		case <-ctx.Done():
			errors <- ctx.Err()
		}
	}
}

func (f *frameClient) WaitForResponses(ctx context.Context, c *websocket.Conn, done chan bool, errors chan error) {
	for {
		var v frameTvResponse
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			errors <- err
		} else {
			if err := f.handleResponse(ctx, &v, done); err != nil {
				errors <- err
			}
		}
	}
}

func (f *frameClient) ToggleArtMode(status string) error {
	data := artModeData{
		Value:   status,
		Request: setArtmodeStatus,
		ID:      f.ID,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	msg := frameTvRequest{
		Method: method,
		Params: frameTvRequestParams{
			To:    to,
			Event: event,
			Data:  string(b),
		},
	}

	bytes, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	f.msgs <- bytes

	return nil
}

func (f *frameClient) RequestSendImage(pathToImage string) error {
	content, err := ioutil.ReadFile(pathToImage)
	if err != nil {
		return errors.Wrap(err, "file read error")
	}

	rand.Seed(time.Now().UnixNano())

	connInfo := connectionData{
		D2dMode:      "socket",
		ConnectionID: rand.Intn(10000000000000000),
		ID:           f.ID,
	}

	connBytes, err := json.Marshal(&connInfo)
	if err != nil {
		return err
	}

	data := sendImageRequest{
		Request:   sendImage,
		FileType:  strings.TrimPrefix(filepath.Ext(pathToImage), "."),
		ImageDate: time.Now().Format("2006:01:02 15:04:05"),
		FileSize:  len(content),
		ConnInfo:  string(connBytes),
	}

	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	msg := frameTvRequest{
		Method: method,
		Params: frameTvRequestParams{
			To:    to,
			Event: event,
			Data:  string(dataBytes),
		},
	}

	bytes, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	f.msgs <- bytes

	return nil
}

func (f *frameClient) handleResponse(ctx context.Context, res *frameTvResponse, done chan bool) error {
	switch res.Event {
	case connectEvent:
		return f.handleConnect(ctx, res)
	case readyEvent:
		done <- true
		return nil
	default:
		fmt.Printf("received: %v\n", res)
	}

	return nil
}

func (f *frameClient) handleConnect(ctx context.Context, res *frameTvResponse) error {
	jsonData, err := json.Marshal(res.Data)
	if err != nil {
		return err
	}

	var r frameTvConnectResponse
	if err = json.Unmarshal(jsonData, &r); err != nil {
		return err
	}

	f.ID = r.ID

	return nil
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
