package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type Profile struct {
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
}

type ProfileResponse struct {
	UserID     string    `json:"user_id,omitempty"`
	Version    string    `json:"version,omitempty"`
	UpdateTime time.Time `json:"update_time"`
}

var protectedWords = []string{"mod", "admin", "system"}

func RpcSaveProfile(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("unauthenticated", 16)
	}

	payloadProfile := Profile{}

	if err := json.Unmarshal([]byte(payload), &payloadProfile); err != nil {
		logger.WithField("user_id", userID).Info("Error unmarshalling payload to JSON: %v", err)
		return "", runtime.NewError("Cannot unmarshal type", 3)
	}

	if payloadProfile == (Profile{}) {
		logger.WithField("user_id", userID).Info("Empty profile payload")
		return "", runtime.NewError("Request is empty, please provide profile fields.", 3)
	}

	// Payload Validation

	// Display Name

	payloadProfile.DisplayName = strings.TrimSpace(payloadProfile.DisplayName)

	for _, r := range payloadProfile.DisplayName {
		if !unicode.IsPrint(r) || r == utf8.RuneError {
			logger.WithField("user_id", userID).Info("Invalid character in display name")
			return "", runtime.NewError("Display name contains invalid characters.", 3)
		}
	}

	if utf8.RuneCountInString(payloadProfile.DisplayName) > 32 || utf8.RuneCountInString(payloadProfile.DisplayName) < 3 {
		logger.WithField("user_id", userID).Info("Display name over the character limit")
		return "", runtime.NewError("Confirm display name is within the acceptable limits 3 - 32 characters.", 3)
	}

	normalizedName := strings.ToLower(payloadProfile.DisplayName)
	for i := 0; i < len(protectedWords); i++ {
		if strings.Contains(normalizedName, protectedWords[i]) {
			logger.WithField("user_id", userID).Info("Protected word in DisplayName")
			return "", runtime.NewError("Invalid Display name please try again.", 3)
		}
	}

	// Bio

	if utf8.RuneCountInString(payloadProfile.Bio) > 190 {
		logger.WithField("user_id", userID).Info("Bio over the character limit")
		return "", runtime.NewError("Character limit exceeded", 3)
	}

	// Avatar validation == too much sidetracking for now will revisit later

	reads := []*runtime.StorageRead{{Collection: "profiles", Key: "profile", UserID: userID}}

	acks, err := nk.StorageWriteRetry(ctx, reads, func(objects []*api.StorageObject) ([]*runtime.StorageWrite, error) {
		profile := Profile{}
		version := "*"
		permissionRead := 2
		permissionWrite := 0
		if len(objects) > 0 {
			if err := json.Unmarshal([]byte(objects[0].Value), &profile); err != nil {
				logger.WithField("user_id", userID).Error("Error unmarshaling read response type to JSON: %v", err)
				return nil, runtime.NewError("unable to unmarshal read content", 13)
			}
			version = objects[0].Version
		}

		profile.DisplayName = payloadProfile.DisplayName
		profile.Bio = payloadProfile.Bio
		profile.Avatar = payloadProfile.Avatar

		value, err := json.Marshal(profile)
		if err != nil {
			logger.WithField("user_id", userID).Error("Error marshaling profile changes to JSON: %v", err)
			return nil, runtime.NewError("unable to marshal profile", 13)
		}

		return []*runtime.StorageWrite{{
			Collection:      "profiles",
			Key:             "profile",
			UserID:          userID,
			Value:           string(value),
			Version:         version,
			PermissionRead:  permissionRead,
			PermissionWrite: permissionWrite,
		}}, nil
	}, 3)

	var runtimeErr *runtime.Error

	if err != nil {
		logger.WithField("user_id", userID).Error("Error saving profile %v", err)
		if errors.As(err, &runtimeErr) {
			return "", runtimeErr

		}
		if errors.Is(err, runtime.ErrStorageWriteExhaustedRetries) {
			return "", runtime.NewError("Profile was updated elsewhere, please try again.", 10)
		}
		return "", runtime.NewError("Error saving profile, please try again.", 13)
	}

	res := &ProfileResponse{
		UserID:     userID,
		Version:    acks[0].Version,
		UpdateTime: acks[0].GetUpdateTime().AsTime(),
	}

	clientRes, err := json.Marshal(res)
	if err != nil {
		logger.WithField("user_id", userID).Error("Error marshaling  type to JSON: %v", err)
		return "", runtime.NewError("unable to marshal response", 13)
	}

	return string(clientRes), nil
}
