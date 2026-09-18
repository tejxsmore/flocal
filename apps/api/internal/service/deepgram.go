package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"

	"flocal/internal/config"
)

type DeepgramWord struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence"`
}

type DeepgramResult struct {
	IsFinal    bool
	Transcript string
	Words      []DeepgramWord
}

var fillerWords = map[string]bool{
	"um": true, "uh": true, "umm": true, "uhh": true, "erm": true, "hmm": true,
}

func IsFillerWord(word string) bool {
	w := strings.ToLower(strings.Trim(word, ".,!?"))
	return fillerWords[w]
}

type DeepgramStreamClient struct {
	conn    *websocket.Conn
	results chan DeepgramResult
	errs    chan error
}

func ConnectDeepgramStream(ctx context.Context, cfg config.DeepgramConfig) (*DeepgramStreamClient, error) {
	q := url.Values{}
	q.Set("model", cfg.Model)
	q.Set("language", cfg.Language)
	q.Set("smart_format", "true")
	q.Set("punctuate", "true")
	q.Set("interim_results", "true")
	q.Set("filler_words", "true")
	q.Set("endpointing", "300")

	dgURL := "wss://api.deepgram.com/v1/listen?" + q.Encode()

	header := http.Header{}
	header.Set("Authorization", "Token "+cfg.APIKey)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, dgURL, header)
	if err != nil {
		return nil, fmt.Errorf("service: connect deepgram stream: %w", err)
	}

	c := &DeepgramStreamClient{
		conn:    conn,
		results: make(chan DeepgramResult, 32),
		errs:    make(chan error, 1),
	}
	go c.readLoop()
	return c, nil
}

func (c *DeepgramStreamClient) readLoop() {
	defer close(c.results)
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.errs <- err
			return
		}

		var raw struct {
			Type    string `json:"type"`
			IsFinal bool   `json:"is_final"`
			Channel struct {
				Alternatives []struct {
					Transcript string         `json:"transcript"`
					Words      []DeepgramWord `json:"words"`
				} `json:"alternatives"`
			} `json:"channel"`
		}
		if err := json.Unmarshal(msg, &raw); err != nil {
			continue
		}
		if raw.Type != "Results" || len(raw.Channel.Alternatives) == 0 {
			continue
		}
		alt := raw.Channel.Alternatives[0]
		if alt.Transcript == "" {
			continue
		}
		c.results <- DeepgramResult{
			IsFinal:    raw.IsFinal,
			Transcript: alt.Transcript,
			Words:      alt.Words,
		}
	}
}

func (c *DeepgramStreamClient) SendAudio(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *DeepgramStreamClient) Results() <-chan DeepgramResult {
	return c.results
}

func (c *DeepgramStreamClient) Close() error {
	_ = c.conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"CloseStream"}`))
	return c.conn.Close()
}
