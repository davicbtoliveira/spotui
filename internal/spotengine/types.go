package spotengine

type EventType string

const (
	EventTypeReady            EventType = "ready"
	EventTypePlaying          EventType = "playing"
	EventTypePaused           EventType = "paused"
	EventTypeStopped          EventType = "stopped"
	EventTypeActive           EventType = "active"
	EventTypeInactive         EventType = "inactive"
	EventTypeMetadata         EventType = "metadata"
	EventTypeVolume           EventType = "volume"
	EventTypeSeek             EventType = "seek"
	EventTypeAuthorizationURL EventType = "authorization_url"
	EventTypeError            EventType = "error"
	EventTypeAccountProduct   EventType = "account_product"
)

type Track struct {
	URI        string
	Name       string
	Artist     string
	Album      string
	DurationMS int
}

type Event struct {
	Type       EventType
	Track      *Track
	ContextURI string
	URI        string
	PositionMS int
	DurationMS int
	Volume     int
	VolumeMax  int
	URL        string
	Err        error
	Product    string
}
