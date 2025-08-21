package domain

import (
	"errors"
	"time"
)

type Token struct {
	Access  string
	Refresh string
}

func NewToken(a, r string) Token {
	return Token{
		Access:  a,
		Refresh: r,
	}
}

func (t *Token) SetAccess(token string) {
	t.Access = token
}

func (t *Token) SetRefresh(token string) {
	t.Refresh = token
}

type Meta struct {
	Exp    time.Time
	UserID string
	Ctx    DeviceCtx
}

type DeviceCtx struct {
	AppId    int32
	DeviceID int32
}

func NewMeta(ttl time.Duration, userId string, appID, deviceId int32) Meta {
	return Meta{
		Exp:    time.Now().Add(ttl),
		UserID: userId,
		Ctx:    NewDeviceCtx(appID, deviceId),
	}
}

func NewDeviceCtx(appID, deviceId int32) DeviceCtx {
	return DeviceCtx{
		AppId:    appID,
		DeviceID: deviceId,
	}
}

type RefreshToken struct {
	Token string
	Meta  Meta
}

func NewRefreshTorken(t string, m Meta) RefreshToken {
	return RefreshToken{
		Token: t,
		Meta:  m,
	}
}

func (dctx DeviceCtx) Compare(otherDctx DeviceCtx) bool {
	return dctx.AppId == otherDctx.AppId && dctx.DeviceID == otherDctx.DeviceID
}

// / implement for tokener.Claims interface
func (m Meta) Claims() map[string]any {
	return map[string]any{
		"user_id":   m.UserID,
		"app_id":    m.Ctx.AppId,
		"device_id": m.Ctx.DeviceID,
		"exp":       m.Exp.Unix(),
	}
}

// / implement for tokenerAdapter.UnClaims interface
func (m *Meta) UnClaims(claims map[string]any) error {
	sub, ok := claims["user_id"].(string)
	if !ok || sub == "" {
		return errors.New("claims: missing or invalid user_id")
	}

	appF, ok := claims["app_id"].(float64)
	if !ok {
		return errors.New("claims: missing or invalid app_id")
	}

	devF, ok := claims["device_id"].(float64)
	if !ok {
		return errors.New("claims: missing or invalid device_id")
	}

	expF, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("claims: missing or invalid exp")
	}

	m.UserID = sub
	m.Ctx = NewDeviceCtx(int32(appF), int32(devF))
	m.Exp = time.Unix(int64(expF), 0)
	return nil
}
