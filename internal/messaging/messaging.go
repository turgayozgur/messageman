package messaging

// Messager interface
type Messager interface {
	EnsureCanConnect(params []interface{}) bool
	Send(mainAPI string, name string, message []byte) error
	Receive(mainAPI string, name string, fn func([]byte) bool) error
}
