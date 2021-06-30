package messaging

import "encoding/json"

type Message struct {
	Body    []byte
	Headers map[string][]byte
}

type Wrapper interface {
	Wrap(body []byte, headers map[string][]byte) (message []byte, err error)
	Unwrap(message []byte) (body []byte, headers map[string][]byte, err error)
}

type DefaultWrapper struct {
}

func (w *DefaultWrapper) Wrap(body []byte, headers map[string][]byte) (message []byte, err error) {
	return json.Marshal(&Message{Body: body, Headers: headers})
}

func (w *DefaultWrapper) Unwrap(message []byte) (body []byte, headers map[string][]byte, err error) {
	var msg Message
	if err = json.Unmarshal(message, &msg); err != nil {
		return nil, nil, err
	}
	return msg.Body, msg.Headers, nil
}
