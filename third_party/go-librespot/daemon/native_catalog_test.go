package daemon

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	librespot "github.com/devgianlu/go-librespot"
	extmetadatapb "github.com/devgianlu/go-librespot/proto/spotify/extendedmetadata"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/spclient"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestNativeArtistHydratesAlbumGroupEntries(t *testing.T) {
	artistGID := make([]byte, 16)
	artistGID[15] = 1
	albumGID := make([]byte, 16)
	albumGID[15] = 2
	artistURI := librespot.SpotifyIdFromGid(librespot.SpotifyIdType("artist"), artistGID).Uri()
	albumURI := librespot.SpotifyIdFromGid(librespot.SpotifyIdType("album"), albumGID).Uri()

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		var batch extmetadatapb.BatchedEntityRequest
		if err := proto.Unmarshal(body, &batch); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		entity := batch.EntityRequest[0]
		kind := entity.Query[0].ExtensionKind
		var message proto.Message
		switch kind {
		case extmetadatapb.ExtensionKind_ARTIST_V4:
			message = &metadatapb.Artist{
				Gid:        artistGID,
				Name:       stringPointer("Arctic Monkeys"),
				AlbumGroup: []*metadatapb.AlbumGroup{{Album: []*metadatapb.Album{{Gid: albumGID}}}},
			}
		case extmetadatapb.ExtensionKind_ALBUM_V4:
			message = &metadatapb.Album{
				Gid:  albumGID,
				Name: stringPointer("AM"),
				Date: &metadatapb.Date{Year: intPointer(2013)},
				Disc: []*metadatapb.Disc{{Track: []*metadatapb.Track{{}}}},
			}
		default:
			t.Errorf("unexpected extension kind: %s", kind)
			return
		}
		payload, err := anypb.New(message)
		if err != nil {
			t.Errorf("pack metadata: %v", err)
			return
		}
		response, err := proto.Marshal(&extmetadatapb.BatchedExtensionResponse{ExtendedMetadata: []*extmetadatapb.EntityExtensionDataArray{{
			ExtensionKind: kind,
			ExtensionData: []*extmetadatapb.EntityExtensionData{{
				Header:        &extmetadatapb.EntityExtensionDataHeader{StatusCode: 200},
				EntityUri:     entity.EntityUri,
				ExtensionData: payload,
			}},
		}}})
		if err != nil {
			t.Errorf("encode response: %v", err)
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writer.Write(response)
	}))
	defer server.Close()

	client, err := spclient.NewSpclient(context.Background(), &librespot.NullLogger{}, server.Client(), func(context.Context) string {
		return server.Listener.Addr().String()
	}, func(context.Context, bool) (string, error) {
		return "token", nil
	}, "device", "client-token")
	if err != nil {
		t.Fatal(err)
	}

	result, err := nativeArtist(context.Background(), client, ApiRequestDataNativeCatalog{URI: artistURI, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	response := result.(map[string]any)
	albums := response["albums"].(map[string]any)["items"].([]any)
	if len(albums) != 1 {
		t.Fatalf("albums: %#v", albums)
	}
	album := albums[0].(map[string]any)
	if album["uri"] != albumURI || album["name"] != "AM" || album["total_tracks"] != 1 {
		t.Fatalf("album summary: %#v", album)
	}
}

func TestNativeAlbumHydratesTrackEntries(t *testing.T) {
	albumGID := make([]byte, 16)
	albumGID[15] = 2
	trackGID := make([]byte, 16)
	trackGID[15] = 3
	albumURI := librespot.SpotifyIdFromGid(librespot.SpotifyIdType("album"), albumGID).Uri()
	trackURI := librespot.SpotifyIdFromGid(librespot.SpotifyIdType("track"), trackGID).Uri()

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		var batch extmetadatapb.BatchedEntityRequest
		if err := proto.Unmarshal(body, &batch); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		entity := batch.EntityRequest[0]
		kind := entity.Query[0].ExtensionKind
		var message proto.Message
		switch kind {
		case extmetadatapb.ExtensionKind_ALBUM_V4:
			message = &metadatapb.Album{
				Gid:  albumGID,
				Name: stringPointer("AM"),
				Date: &metadatapb.Date{Year: intPointer(2013)},
				Disc: []*metadatapb.Disc{{Track: []*metadatapb.Track{{Gid: trackGID}}}},
			}
		case extmetadatapb.ExtensionKind_TRACK_V4:
			message = &metadatapb.Track{
				Gid:        trackGID,
				Name:       stringPointer("Star Treatment"),
				Album:      &metadatapb.Album{Gid: albumGID, Name: stringPointer("AM")},
				Duration:   intPointer(250000),
				Number:     intPointer(1),
				DiscNumber: intPointer(1),
			}
		default:
			t.Errorf("unexpected extension kind: %s", kind)
			return
		}
		payload, err := anypb.New(message)
		if err != nil {
			t.Errorf("pack metadata: %v", err)
			return
		}
		response, err := proto.Marshal(&extmetadatapb.BatchedExtensionResponse{ExtendedMetadata: []*extmetadatapb.EntityExtensionDataArray{{
			ExtensionKind: kind,
			ExtensionData: []*extmetadatapb.EntityExtensionData{{
				Header:        &extmetadatapb.EntityExtensionDataHeader{StatusCode: 200},
				EntityUri:     entity.EntityUri,
				ExtensionData: payload,
			}},
		}}})
		if err != nil {
			t.Errorf("encode response: %v", err)
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writer.Write(response)
	}))
	defer server.Close()

	client, err := spclient.NewSpclient(context.Background(), &librespot.NullLogger{}, server.Client(), func(context.Context) string {
		return server.Listener.Addr().String()
	}, func(context.Context, bool) (string, error) {
		return "token", nil
	}, "device", "client-token")
	if err != nil {
		t.Fatal(err)
	}

	result, err := nativeAlbum(context.Background(), client, ApiRequestDataNativeCatalog{URI: albumURI, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	response := result.(map[string]any)
	tracks := response["tracks"].(map[string]any)["items"].([]any)
	if len(tracks) != 1 {
		t.Fatalf("tracks: %#v", tracks)
	}
	track := tracks[0].(map[string]any)
	if track["uri"] != trackURI || track["name"] != "Star Treatment" || track["duration_ms"] != int32(250000) {
		t.Fatalf("track metadata: %#v", track)
	}
}

func TestTrackValuePreservesProvidedURI(t *testing.T) {
	metadataGID := make([]byte, 16)
	metadataGID[15] = 4
	providedGID := make([]byte, 16)
	providedGID[15] = 5
	providedURI := librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, providedGID).Uri()

	value := trackValue(&metadatapb.Track{Gid: metadataGID, Name: stringPointer("Track")}, providedURI, "")
	if value["uri"] != providedURI {
		t.Fatalf("track URI: got %v, want %s", value["uri"], providedURI)
	}
}

func stringPointer(value string) *string { return &value }
func intPointer(value int32) *int32      { return &value }
