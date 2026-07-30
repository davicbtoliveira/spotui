package daemon

import (
	"testing"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"google.golang.org/protobuf/proto"
)

func TestSearchContextURIUsesNativeResolver(t *testing.T) {
	if got := searchContextURI("Daft Punk"); got != "spotify:search:Daft+Punk" {
		t.Fatalf("context URI = %q", got)
	}
}

func TestSearchTrackResponseTranslatesMetadata(t *testing.T) {
	response := searchTrackResponse("spotify:track:hello", &metadatapb.Track{
		Name:     proto.String("Hello"),
		Duration: proto.Int32(295000),
		Artist: []*metadatapb.Artist{
			{Name: proto.String("Adele")},
		},
		Album: &metadatapb.Album{Name: proto.String("25")},
	})

	if response.Uri != "spotify:track:hello" ||
		response.Name != "Hello" ||
		len(response.ArtistNames) != 1 ||
		response.ArtistNames[0] != "Adele" ||
		response.AlbumName != "25" ||
		response.Duration != 295000 {
		t.Fatalf("response = %#v", response)
	}
}
