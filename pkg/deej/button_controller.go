package deej

type ButtonController interface {
	Start() error
	Stop()
	SubscribeToButtonStateChangeEvents() chan ButtonStateChangeEvent
}

// ButtonStateChangeEvent represents an event indicating a change in button state
type ButtonStateChangeEvent struct {
	ButtonID int // ID of the button
	State    int // New state of the button (0 for released, 1 for pressed)
}
