package health

type Readiness interface {
	Name() string
	Validate() error

	EventTypes() []interface{}
	EventsReceiver(event interface{})
}
