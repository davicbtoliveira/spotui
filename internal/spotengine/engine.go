package spotengine

import "context"

type Engine interface {
	Start(context.Context) error
	SearchTracks(context.Context, SearchRequest) (SearchPage, error)
	Play(context.Context, string) error
	Pause(context.Context) error
	Resume(context.Context) error
	Next(context.Context) error
	Previous(context.Context) error
	SetVolume(context.Context, int) error
	SetAutoplay(context.Context, bool) error
	Events() <-chan Event
	Close(context.Context) error
}

type SearchRequest struct {
	Query  string
	Offset int
	Limit  int
}

type SearchPage struct {
	Tracks []Track
	Total  int
	Offset int
}
