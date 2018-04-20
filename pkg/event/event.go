// Package event mediates notification between controllers and recorder
package event

// Action represents the kind of object change we're notifying
type Action int

const (
	// Delete is the object deletion Action
	Delete Action = iota

	// Upsert is the update or create Action
	Upsert
)

// Notification conveys an object delete/upsert notification
type Notification struct {
	Action Action
	Key    string
	Kind   string
	Object []byte
}

// Notifier mediates notifications between controllers and recorder
type Notifier interface {
	Send(notif *Notification)
	ReadChan() <-chan Notification
}

// Unbuffered implements Notifier
type Unbuffered struct {
	c chan Notification
}

// New creates an Unbuffered
func New() *Unbuffered {
	return &Unbuffered{
		c: make(chan Notification),
	}
}

// Send sends a notification
func (n *Unbuffered) Send(notif *Notification) {
	n.c <- *notif
}

// ReadChan returns a channel to read Notifications from
func (n *Unbuffered) ReadChan() <-chan Notification {
	return n.c
}
