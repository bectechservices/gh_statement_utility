package requests

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gookit/validate"
)

//FormRequest form request type
type FormRequest interface {
	Validate(context buffalo.Context) (*validate.Validation, error)
	GetBoundValue(c buffalo.Context) interface{}
}
