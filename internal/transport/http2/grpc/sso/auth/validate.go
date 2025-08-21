package grpctransportauth

import (
	"context"
	"errors"

	"github.com/eragon-mdi/go-playground/validator"
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
)

func validate(v any) error {
	targedRequestStruct, err := reqToInternalValidateStruct(v)
	if err != nil {
		return err
	}

	return validator.Validate(context.Background(), targedRequestStruct)
}

func reqToInternalValidateStruct(v any) (any, error) {
	switch t := v.(type) {
	case *sso.RegisterRequest:
		if t.User == nil {
			return nil, errors.New("user is required")
		}
		return RegisterReqValidation{
			UserValidation: newUserTovalidate(t.User),
		}, nil

	case *sso.LoginRequest:
		if t.User == nil {
			return nil, errors.New("user is required")
		}
		if t.Ctx == nil {
			return nil, errors.New("device context is required")
		}
		return LoginReqValidation{
			UserValidation:      newUserTovalidate(t.User),
			DeviceCtxValidation: newDeviceCtxTovalidate(t.Ctx),
		}, nil

	case *sso.LogoutRequest:
		if t.Ctx == nil {
			return nil, errors.New("device context is required")
		}
		return LogoutReqValidation{
			RefreshTokenValidate: RefreshTokenValidate{
				Refresh: t.Refresh,
			},
			DeviceCtxValidation: newDeviceCtxTovalidate(t.Ctx),
		}, nil

	case *sso.RefreshRequest:
		if t.Ctx == nil {
			return nil, errors.New("device context is required")
		}
		return RefreshReqValidation{
			RefreshTokenValidate: RefreshTokenValidate{
				Refresh: t.Refresh,
			},
			DeviceCtxValidation: newDeviceCtxTovalidate(t.Ctx),
		}, nil

	default:
		return nil, errors.New("bad request type")
	}
}

func newUserTovalidate(u *sso.User) UserValidation {
	return UserValidation{
		Email:    u.Email,
		Password: u.Password,
	}
}

func newDeviceCtxTovalidate(ctx *sso.DeviceContext) DeviceCtxValidation {
	return DeviceCtxValidation{
		AppId:    ctx.AppId,
		DeviceId: ctx.DeviceId,
	}
}
