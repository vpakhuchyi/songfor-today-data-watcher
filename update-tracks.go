package updatetracks

import (
	"context"
	"os"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-resty/resty/v2"
	"github.com/vpakhuchyi/songfor-today/adapters/deezer"
	"github.com/vpakhuchyi/songfor-today/adapters/firestore"
	"golang.org/x/exp/slog"
)

var (
	projectID         = os.Getenv("PROJECT_ID")
	playlistID        = os.Getenv("PLAYLIST_ID")
	deezerAccessToken = os.Getenv("DEEZER_ACCESS_TOKEN")

	deezerClient    deezer.Client
	firestoreClient firestore.Adapter
)

type MessagePublishedData struct {
	Message PubSubMessage
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func init() {
	ctx := context.Background()

	fs, err := gcpfirestore.NewClient(ctx, projectID)
	if err != nil {
		slog.ErrorCtx(ctx, "Failed to create client", "err", err)

		return
	}

	deezerClient = deezer.New(resty.New())
	firestoreClient = firestore.New(fs)

	functions.CloudEvent("UpdateTracks", UpdateTracks)
}

func UpdateTracks(ctx context.Context, e event.Event) error {
	slog.InfoCtx(ctx, "UpdateTracks lambda has been triggered")

	tracks, err := deezerClient.GetPlaylistTracks(ctx, deezerAccessToken, playlistID)
	if err != nil {
		slog.ErrorCtx(ctx, "Failed to fetch tracks from deezer", "err", err)

		return err
	}

	err = firestoreClient.PutTracks(ctx, tracks)
	if err != nil {
		slog.ErrorCtx(ctx, "Failed to update tracks in datastore", "err", err)

		return err
	}

	slog.InfoCtx(ctx, "Tracks have been successfully updated", "total count", len(tracks))

	return nil
}
