package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

var currencyName string = "coins"
var starterCoins int64 = 100

func AfterAuthenticateDevice(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateDeviceRequest) error {
	return initializeNewPlayer(ctx, logger, nk, out)
}

func initializeNewPlayer(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, out *api.Session) error {
	if !out.GetCreated() {
		return nil
	}

	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		logger.Error("UserID is missing")
		return runtime.NewError("unauthenticated", 16)
	}

	log := logger.WithField("user_id", userID)

	profile := Profile{
		DisplayName: "Player_" + userID[0:8],
		Avatar:      "avatar_01",
		Bio:         "",
	}

	err := ProfileValidation(profile, log)
	if err != nil {
		log.Error("Profile does not pass validation %v", err)
		return err
	}

	payload, err := json.Marshal(profile)
	if err != nil {
		log.Error("error marshaling profile %v", err)
		return err
	}

	profileWrite := []*runtime.StorageWrite{{
		Collection:      collection,
		Key:             key,
		UserID:          userID,
		Value:           string(payload),
		PermissionRead:  2,
		PermissionWrite: 0,
		Version:         "*",
	}}

	walletWrite := []*runtime.WalletUpdate{{
		UserID:    userID,
		Changeset: map[string]int64{currencyName: starterCoins},
		Metadata:  map[string]interface{}{"reason": "starter_grant"},
	}}

	_, res, err := nk.MultiUpdate(ctx, nil, profileWrite, nil, walletWrite, true)
	if err != nil {
		log.Error("Error before errors Is", err)
		if errors.Is(err, runtime.ErrStorageRejectedVersion) {
			updCoins, _, err := nk.WalletUpdate(ctx, walletWrite[0].UserID, walletWrite[0].Changeset, walletWrite[0].Metadata, true)
			if err != nil {
				log.Error("error granting player %v starter %v, error: %v", starterCoins, currencyName, err)
				return err
			}
			log.Info("Initial coins grant successful, new balance is: %v", updCoins[currencyName])
			return nil
		}
		log.Error("error setting up player profile %v", err)
		return err
	}

	log.Info("Initial coins grant successful, new balance is: %v", res[0].Updated[currencyName])

	return nil
}
