package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	err := initializer.RegisterRpc("healthcheck", RpcHealthcheck)
	if err != nil {
		return err
	}
	err = initializer.RegisterRpc("save_profile", RpcSaveProfile)
	if err != nil {
		return err
	}
	err = initializer.RegisterRpc("get_profile", RpcGetProfile)
	if err != nil {
		return err
	}
	err = initializer.RegisterBeforeWriteStorageObjects(BeforeClientWriteRequest)
	if err != nil {
		return err
	}
	err = initializer.RegisterAfterAuthenticateDevice(AfterAuthenticateDevice)
	if err != nil {
		return err
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())
	return nil
}
