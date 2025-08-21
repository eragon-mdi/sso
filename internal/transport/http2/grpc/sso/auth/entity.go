package grpctransportauth

import (
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/domain"
)

type UserValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=12"`
}

type DeviceCtxValidation struct {
	AppId    int32 `validate:"required,gt=0"`
	DeviceId int32 `validate:"required,gt=0"`
}

type RefreshTokenValidate struct {
	Refresh string `validate:"required"`
}

type RegisterReqValidation struct {
	UserValidation
}

type LoginReqValidation struct {
	UserValidation
	DeviceCtxValidation
}

type LogoutReqValidation struct {
	RefreshTokenValidate
	DeviceCtxValidation
}

type RefreshReqValidation struct {
	RefreshTokenValidate
	DeviceCtxValidation
}

func userFromRegisterReq(req *sso.RegisterRequest) domain.User {
	return domain.User{
		Email:    req.GetUser().GetEmail(),
		Password: req.GetUser().GetPassword(),
	}
}

func userFromLoginReq(req *sso.LoginRequest) domain.User {
	return domain.User{
		Email:    req.GetUser().GetEmail(),
		Password: req.GetUser().GetPassword(),
	}
}

func deviceCtxFromReq(reqDeviceCtx *sso.DeviceContext) domain.DeviceCtx {
	return domain.NewDeviceCtx(reqDeviceCtx.AppId, reqDeviceCtx.DeviceId)
}
