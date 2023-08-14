package openai

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	utils "github.com/sashabaranov/go-openai/internal"
)

var (
	errorPrefix = []byte(`data: {"error":`)
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
)

type streamable interface {
	GlmChatCompletionStreamResponseResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler
}

type Event struct {
	ID    []byte
	Data  []byte
	Event []byte
}

var pool = sync.Pool{
	New: func() interface{} {
		fmt.Println("********* get event pool *********")
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

	event := pool.Get().(*Event)

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
		if e != nil {
			if e.Event != nil {
				if bytes.Compare(e.Event, []byte("finish")) == 0 {
					stream.isFinished = true
				}
				event.Event = e.Event
			}
			if e.ID != nil {
				event.ID = e.ID
			}
			if e.Data != nil {
				event.Data = append(event.Data[:], e.Data...)
			}
		}
		if e == nil {
			response = T{
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
			}

			putEvent(event)
			fmt.Printf("----response: %+v\n", response)

			return response, nil
		}

	}
}

func putEvent(e *Event) {
	e.ID = nil
	e.Event = nil
	e.Data = nil
	pool.Put(e)
	fmt.Println("========== put event pool ===========")
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
		switch {
		case bytes.HasPrefix(line, headerID):
			e.ID = append([]byte(nil), trimHeader(len(headerID), line)...)
		case bytes.HasPrefix(line, headerData):
			// The spec allows for multiple data fields per event, concatenated them with "\n".
			e.Data = append(e.Data[:], append(trimHeader(len(headerData), line), byte('\n'))...)
		// The spec says that a line that simply contains the string "data" should be treated as a data field with an empty body.
		case bytes.Equal(line, bytes.TrimSuffix(headerData, []byte(":"))):
			e.Data = append(e.Data, byte('\n'))
		case bytes.HasPrefix(line, headerEvent):
			e.Event = append([]byte(nil), trimHeader(len(headerEvent), line)...)
		// case bytes.HasPrefix(line, headerRetry):
		// 	e.Retry = append([]byte(nil), trimHeader(len(headerRetry), line)...)
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
