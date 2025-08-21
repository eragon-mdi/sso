package grpctransportpermission

import (
	"github.com/eragon-mdi/protos/gen/go/sso/v1"
	"github.com/eragon-mdi/sso/internal/domain"
)

type IsAdminReqValidation struct {
	UserId string `validate:"required,uuid4"`
}

func userFromIsAdminReq(req *sso.IsAdminRequest) domain.User {
	return domain.User{
		ID: req.GetUserId(),
	}
}
