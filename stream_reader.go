package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	utils "github.com/gtkit/zhipuAi/internal"
)

var (
	errorPrefix = []byte(`data: {"error":`)
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
	headerMeta  = []byte("meta:")
)

type streamable interface {
	GlmChatCompletionStreamResponseResponse
}

type streamReader[T streamable] struct {
	isFinished bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler
}

type Event struct {
	ID    []byte
	Data  []byte
	Event []byte
	Meta  []byte
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Event{}
	},
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

func (stream *streamReader[T]) processLines() (T, error) {
	var (
		hasErrorPrefix bool
		response       T
	)

	event, ok := pool.Get().(*Event)
	if !ok {
		return *new(T), nil
	}

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')

		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return *new(T), fmt.Errorf("error, %w", respErr.Error)
			}
			return *new(T), readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)

		if bytes.HasPrefix(noSpaceLine, errorPrefix) {
			hasErrorPrefix = true
		}

		e, _ := processEvent(noSpaceLine)

		if e == nil {
			meta := &GlmMeta{}
			if len(event.Meta) > 0 {
				err := json.Unmarshal(event.Meta, meta)
				if err != nil {
					log.Println("---Meta Unmarshal error:", err)
				}
			}

			response = T(GlmChatCompletionStreamResponseResponse{
				ID:    string(event.ID),
				Event: string(event.Event),
				Data:  string(event.Data),
				Choices: []ChatCompletionStreamChoice{
					{
						Delta: ChatCompletionStreamChoiceDelta{
							Content: string(event.Data),
						},
					},
				},
				Meta: *meta,
			})
			putEvent(event)

			return response, nil
		}

		if e.Event != nil {
			if bytes.Equal(e.Event, []byte("finish")) {
				stream.isFinished = true
			}
			event.Event = e.Event
		}
		if e.ID != nil {
			event.ID = e.ID
		}
		if e.Data != nil {
			event.Data = e.Data
		}
		if e.Meta != nil {
			event.Meta = e.Meta
		}
	}
}

func putEvent(e *Event) {
	e.ID = nil
	e.Event = nil
	e.Data = nil
	pool.Put(e)
}

func (stream *streamReader[T]) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}

func processEvent(msg []byte) (event *Event, err error) {
	var e Event

	if len(msg) < 1 {
		return nil, errors.New("event message was empty")
	}

	// Normalize the crlf to lf to make it easier to split the lines.
	// Split the line by "\n" or "\r", per the spec.
	for _, line := range bytes.FieldsFunc(msg, func(r rune) bool { return r == '\n' || r == '\r' }) {
		// fmt.Println("line:", string(line))
		switch {
		case bytes.HasPrefix(line, headerID):
			e.ID = append([]byte(nil), trimHeader(len(headerID), line)...)
		case bytes.HasPrefix(line, headerData):
			// The spec allows for multiple data fields per event, concatenated them with "\n".
			e.Data = append(e.Data, append(trimHeader(len(headerData), line), byte('\n'))...)
		// The spec says that a line that simply contains the string "data" should be treated
		// as a data field with an empty body.
		case bytes.Equal(line, bytes.TrimSuffix(headerData, []byte(":"))):
			e.Data = append(e.Data, byte('\n'))
		case bytes.HasPrefix(line, headerEvent):
			e.Event = append([]byte(nil), trimHeader(len(headerEvent), line)...)
		case bytes.HasPrefix(line, headerMeta):
			e.Meta = append([]byte(nil), trimHeader(len(headerMeta), line)...)
			// fmt.Println("-----meta:", string(e.Meta))
		default:
			// Ignore any garbage that doesn't match what we're looking for.
		}
	}

	// Trim the last "\n" per the spec.
	e.Data = bytes.TrimSuffix(e.Data, []byte("\n"))

	return &e, err
}
func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}

	data = data[size:]
	// Remove optional leading whitespace
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	// Remove trailing new line
	if len(data) > 0 && data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}
	return data
}
