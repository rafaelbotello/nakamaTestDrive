package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"uuid"

	"github.com/heroiclabs/nakama-common/runtime"
)

type GetProfileRequest struct {
	UserID string `json:"user_id"`
}

type GetProfileResponse struct {
	UserID string `json:"user_id"`
	Profile
}

func RpcGetProfile(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("unauthenticated", 16)
	}

	reqPayload := GetProfileRequest{}

	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &reqPayload); err != nil {
			logger.WithField("user_id", userID).Info("Error unmarshalling payload to JSON: %v", err)
			return "", runtime.NewError("Cannot unmarshal type", 3)
		}
	}

	reqPayload.UserID = strings.TrimSpace(reqPayload.UserID)

	if reqPayload.UserID == "" {
		logger.WithField("user_id", userID).Info("Requesting own profile.")
		reqPayload.UserID = userID
	}

	if _, err := uuid.Parse(reqPayload.UserID); err != nil {
		logger.WithField("user_id", userID).Info("request user_id is invalid.")
		return "", runtime.NewError("user_id is invalid, please provide a valid id.", 3)
	}

	reads := []*runtime.StorageRead{{Collection: "profiles", Key: "profile", UserID: reqPayload.UserID}}

	res, err := nk.StorageRead(ctx, reads)
	if err != nil {
		logger.WithField("user_id", userID).Error("Error reading object: %v", err)
		return "", runtime.NewError("there was an error during the search, please try again.", 13)
	}

	profile := Profile{}

	if len(res) >= 1 {
		if res[0].PermissionRead != 2 && res[0].UserId != userID {
			logger.WithField("user_id", userID).Info("User unauthorized")
			return "", runtime.NewError("profile not found", 5)
		}

		if err := json.Unmarshal([]byte(res[0].Value), &profile); err != nil {
			logger.WithField("user_id", userID).Error("Error unmarshaling read response type to JSON: %v", err)
			return "", runtime.NewError("unable to unmarshal read content", 13)
		}
	} else {
		logger.WithField("user_id", userID).Info("Profile not found, read came back empty")
		return "", runtime.NewError("profile not found", 5)
	}

	clientRes, err := json.Marshal(GetProfileResponse{UserID: res[0].UserId, Profile: profile})
	if err != nil {
		logger.WithField("user_id", userID).Error("Error marshaling  type to JSON: %v", err)
		return "", runtime.NewError("unable to marshal response", 13)
	}

	return string(clientRes), nil
}
