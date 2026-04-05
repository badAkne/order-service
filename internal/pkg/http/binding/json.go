package binding

import (
	"net/http"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/badAkne/order-service/internal/pkg/http/httph"
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "JSON"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return entity.ErrIncorrectParameters
	}

	if err := httph.DecodeJSON(req, obj); err != nil {
		return err
	}

	return validate(obj)
}
