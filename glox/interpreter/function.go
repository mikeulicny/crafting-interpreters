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
	declaration   ast.FunctionStmt
	closure       *env.Environment
	isInitializer bool
}

func (f function) arity() int {
	return len(f.declaration.Params)
}

func (f function) call(i *Interpreter, arguments []any) (returnVal any) {
	defer func() {
		if err := recover(); err != nil {
			// Catch special error throw of type Return for function returns
			if v, ok := err.(Return); ok {
				// make init methods always return "this"
				if f.isInitializer {
					f.closure.GetAt(0, "this")
					return
				}
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

func (f function) bind(in *instance) function {
	closure := env.New(f.closure)
	closure.Define("this", in)
	return function{
		declaration:   f.declaration,
		closure:       closure,
		isInitializer: f.isInitializer,
	}
}
