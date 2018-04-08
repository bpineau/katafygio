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
	Object string
}

// Notifier mediates notifications between controllers and recorder
type Notifier struct {
	C chan Notification
}

// New creates a new event.Notifier
func New() *Notifier {
	return &Notifier{
		C: make(chan Notification),
	}
}

// Send sends a notification
func (n *Notifier) Send(notif *Notification) {
	n.C <- *notif
}
