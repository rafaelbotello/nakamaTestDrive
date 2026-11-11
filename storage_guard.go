package main

import (
	"context"
	"database/sql"
	"slices"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

var clientWritableCollections = []string{}

func BeforeClientWriteRequest(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.WriteStorageObjectsRequest) (*api.WriteStorageObjectsRequest, error) {
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	for _, object := range in.GetObjects() {
		if !slices.Contains(clientWritableCollections, object.GetCollection()) {
			logger.WithFields(map[string]interface{}{"user_id": userID, "collection": object.GetCollection()}).Warn("Client attempting to write on unauthorized collection")
			return nil, runtime.NewError("Unauthorized", 7)
		}
	}

	return in, nil
}
