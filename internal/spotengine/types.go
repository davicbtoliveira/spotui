package spotengine

type EventType string

const EventTypeMetadata EventType = "metadata"

type Track struct {
	URI        string
	Name       string
	Artist     string
	Album      string
	DurationMS int
}

type Event struct {
	Type  EventType
	Track *Track
}
