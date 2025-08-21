package grpctransportpermission

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

func reqToInternalValidateStruct(v any) (validateStruct any, err error) {
	switch t := v.(type) {
	case *sso.IsAdminRequest:
		validateStruct = IsAdminReqValidation{
			UserId: t.UserId,
		}
	default:
		err = errors.New("bad request")
	}

	return
}
