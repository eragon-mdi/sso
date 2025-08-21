package redisrepo

import (
	"time"

	"github.com/eragon-mdi/sso/internal/domain"
)

type tokenMeta struct {
	UserID   string    `json:"user_id"`
	AppID    int32     `json:"app_id"`
	DeviceID int32     `json:"device_id"`
	Exp      time.Time `json:"exp"`
}

func newTokenMeta(m domain.Meta) *tokenMeta {
	return &tokenMeta{
		UserID:   m.UserID,
		AppID:    m.Ctx.AppId,
		DeviceID: m.Ctx.DeviceID,
		Exp:      m.Exp,
	}
}
