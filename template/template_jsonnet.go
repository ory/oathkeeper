package template

import (
	"github.com/google/go-jsonnet/ast"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/google/go-jsonnet"
)

type JsonNet struct{}

func (j *JsonNet) Render(template string, engine RenderEngine, session *authn.AuthenticationSession, opts renderOptions) (string, error) {

	vm := jsonnet.MakeVM()
	vm.NativeFunction()

	var jsonToString = &NativeFunction{
		Name:   "jsonToString",
		Params: ast.Identifiers{"x"},
		Func: func(x []interface{}) (interface{}, error) {
			bytes, err := json.Marshal(x[0])
			if err != nil {
				return nil, err
			}
			return string(bytes), nil
		},
	}
	vm.NativeFunction(jsonToString)

}
