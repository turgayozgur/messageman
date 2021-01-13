package messaging

// Messager interface
type Messager interface {
	EnsureCanConnect() bool
	NotifyRecover(chan string) chan string
	Queue(service string, name string, message []byte) error
	Work(service string, name string, callback func([]byte) bool) error
	Publish(service string, name string, message []byte) error
	Subscribe(service string, name string, callback func([]byte) bool) error
}
