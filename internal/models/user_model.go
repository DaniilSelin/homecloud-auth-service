package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	PasswordHash        string     `json:"-"`
	IsActive            bool       `json:"is_active"`
	IsEmailVerified     bool       `json:"is_email_verified"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	StorageQuota        int64      `json:"storage_quota"`
	UsedSpace           int64      `json:"used_space"`
	Role                string     `json:"role"`
	IsAdmin             bool       `json:"is_admin"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Методы для работы с пользователем
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

func (u *User) CanLogin() bool {
	return u.IsActive && !u.IsLocked()
}

func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= 5 {
		lockTime := time.Now().Add(15 * time.Minute)
		u.LockedUntil = &lockTime
	}
}

func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}