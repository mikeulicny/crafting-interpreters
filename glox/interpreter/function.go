package interpreter

import (
	"glox/ast"
	"glox/env"
)

type callable interface {
	arity() int
	call(i *Interpreter, arguments []any) any
}

type function struct {
	declaration ast.FunctionStmt
	closure     *env.Environment
}

func (f function) arity() int {
	return len(f.declaration.Params)
}

func (f function) call(i *Interpreter, arguments []any) (returnVal any) {
	defer func() {
		if err := recover(); err != nil {
			// Catch special error throw of type Return for function returns
			if v, ok := err.(Return); ok {
				returnVal = v.Value
				return
			}
			panic(err)
		}
	}()

	environment := env.New(f.closure)
	for idx, param := range f.declaration.Params {
		environment.Define(param.Lexeme, arguments[idx])
	}

	i.executeBlock(f.declaration.Body, environment)
	return nil
}

func (f function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}
