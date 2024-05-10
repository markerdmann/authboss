package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGenerateConfirmCreds(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t)
	key := "test-secret-key"

	tests := []struct {
		name        string
		userID      int
		email       string
		wantErr     bool
		wantUserID  int
		wantEmail   string
		wantCodeLen int
	}{
		{
			name:        "success",
			userID:      123,
			email:       "test@test.com",
			wantUserID:  123, 
			wantEmail:   "test@test.com",
			wantCodeLen: 32,
		},
		{
			name:    "missing userID",
			userID:  0,
			email:   "test@test.com", 
			wantErr: true,
		},
		{
			name:    "missing email",
			userID:  123,
			email:   "",
			wantErr: true,
		},
		{
			name:    "invalid email",
			userID:  123,
			email:   "notanemail",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmCreds, err := GenerateConfirmCreds(ctx, db, key, tt.userID, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateConfirmCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if confirmCreds.UserID != tt.wantUserID {
				t.Errorf("confirmCreds.UserID = %v, want %v", confirmCreds.UserID, tt.wantUserID)
			}
			if confirmCreds.Email != tt.wantEmail {
				t.Errorf("confirmCreds.Email = %v, want %v", confirmCreds.Email, tt.wantEmail)
			}
			if len(confirmCreds.Code) != tt.wantCodeLen {
				t.Errorf("len(confirmCreds.Code) = %v, want %v", len(confirmCreds.Code), tt.wantCodeLen)
			}
		})
	}
}

func TestGenerateConfirmCreds_CreateConfirmCredsError(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t)
	key := "test-secret-key"
	userID := 123
	email := "test@test.com"

	// Simulate an error from CreateConfirmCreds
	oldCreateConfirmCreds := CreateConfirmCreds
	defer func() { CreateConfirmCreds = oldCreateConfirmCreds }()
	CreateConfirmCreds = func(context.Context, dbutil.DB, string, int, string) (*ConfirmCreds, error) {
		return nil, errors.New("CreateConfirmCreds error")
	}

	_, err := GenerateConfirmCreds(ctx, db, key, userID, email)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestGenerateConfirmCreds_VerifyFields(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t)
	key := "test-secret-key"
	userID := 123
	email := "test@test.com"

	confirmCreds, err := GenerateConfirmCreds(ctx, db, key, userID, email)
	if err != nil {
		t.Fatal(err)
	}

	if confirmCreds.UserID != userID {
		t.Errorf("UserID = %v, want %v", confirmCreds.UserID, userID)
	}
	if confirmCreds.Email != email {
		t.Errorf("Email = %v, want %v", confirmCreds.Email, email)
	}
	if confirmCreds.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero value")
	}
	if confirmCreds.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt is in the past")
	}
	if len(confirmCreds.Code) != 32 {
		t.Errorf("len(Code) = %v, want 32", len(confirmCreds.Code))
	}
}