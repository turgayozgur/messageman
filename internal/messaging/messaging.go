package messaging

import (
	"bytes"
	"context"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

// Messager interface
type Messager interface {
	EnsureCanConnect() bool
	NotifyRecover(chan string) chan string
	Queue(service string, name string, message []byte) error
	Work(service string, name string, callback func([]byte) bool) error
	Publish(service string, name string, message []byte) error
	Subscribe(service string, name string, callback func([]byte) bool) error
}

func doRest(wrapper Wrapper, client *http.Client, url string, body []byte) (*http.Response, error) {
	body, headers, err := wrapper.Unwrap(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, string(v))
	}
	req.Header.Set("Content-Type", ContentType)
	return client.Do(req)
}

func doGRPC(wrapper Wrapper, body []byte, fn func(ctx context.Context) error) error {
	body, headers, err := wrapper.Unwrap(body)
	if err != nil {
		return err
	}
	h := make(map[string]string, len(headers))
	for k, v := range headers {
		h[k] = string(v)
	}
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*60))
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(h))
	return fn(ctx)
}
