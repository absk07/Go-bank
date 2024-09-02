// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Account struct {
	ID        int64              `json:"id"`
	Owner     string             `json:"owner"`
	Balance   int64              `json:"balance"`
	Currency  string             `json:"currency"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type Entry struct {
	ID        int64              `json:"id"`
	AccountID int64              `json:"account_id"`
	Amount    int64              `json:"amount"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type Session struct {
	ID           uuid.UUID          `json:"id"`
	Username     string             `json:"username"`
	RefreshToken string             `json:"refresh_token"`
	UserAgent    string             `json:"user_agent"`
	ClientIp     string             `json:"client_ip"`
	IsBlocked    bool               `json:"is_blocked"`
	ExpiresAt    pgtype.Timestamptz `json:"expires_at"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

type Transfer struct {
	ID            int64              `json:"id"`
	FromAccountID int64              `json:"from_account_id"`
	ToAccountID   int64              `json:"to_account_id"`
	Amount        int64              `json:"amount"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

type User struct {
	Username          string             `json:"username"`
	Password          string             `json:"password"`
	Fullname          string             `json:"fullname"`
	Email             string             `json:"email"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	CreatedAt         pgtype.Timestamptz `json:"created_at"`
	IsEmailVerified   bool               `json:"is_email_verified"`
}

type VerifyEmail struct {
	ID         int64              `json:"id"`
	Username   string             `json:"username"`
	Email      string             `json:"email"`
	SecretCode string             `json:"secret_code"`
	IsUsed     bool               `json:"is_used"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	ExpiredAt  pgtype.Timestamptz `json:"expired_at"`
}
