package binding

import (
	"net/http"
)

const (
	MIMEJSON              = "application/json"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

type Binding interface {
	Name() string
	Bind(r *http.Request, obj any) error
}

type BindingUri interface {
	Name() string
	BindUri(params map[string][]string, obj any) error
}

type StructValidator interface {
	ValidateStruct(obj any) error
	Engine() any
}

var Validator StructValidator = &defaultValidator{}

var (
	bJSON         = jsonBinding{}
	bQuery        = queryBinding{}
	bOnlyValidate = onlyValidateBinding{}
)

func validate(obj any) error {
	if Validator == nil {
		return nil
	}

	return Validator.ValidateStruct(obj)
}

type onlyValidateBinding struct{}

func (onlyValidateBinding) Name() string {
	return "OnlyValidate"
}

func (onlyValidateBinding) Bind(_ *http.Request, obj any) error {
	return validate(obj)
}
