package interpreter

import (
	"fmt"
	"glox/ast"
)

// A Class type that identifies behavior of a Class in glox
// This is where the methods of a Class are stored
type Class struct {
	Name    string
	Methods map[string]function
}

func NewClass(name string) *Class {
	return &Class{Name: name}
}

func (c Class) String() string {
	return c.Name
}

func (c Class) call(interpreter *Interpreter, arguments []any) any {
	instance := &instance{class: c}
	initializer := c.findMethod("init")
	if initializer != nil {
		initializer.bind(instance).call(interpreter, arguments)
	}

	return instance
}

func (c Class) arity() int {
	initializer := c.findMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.arity()
}

func (c Class) findMethod(name string) *function {
	if method, ok := c.Methods[name]; ok {
		return &method
	}

	return nil
}

type Instance interface {
	Get(name ast.Token) (any, error)
}

// An instance of a Class in the glox language
// This is where the state of a class is stored
type instance struct {
	class  Class
	fields map[string]any
}

func (i instance) String() string {
	return i.class.Name + " instance"
}

func (i *instance) Get(name ast.Token) (any, error) {
	if val, exists := i.fields[name.Lexeme]; exists {
		return val, nil
	}

	method := i.class.findMethod(name.Lexeme)
	if method != nil {
		return method.bind(i), nil
	}

	return nil, runtimeError{token: name, msg: fmt.Sprintf("Undefined property '%s'.", name.Lexeme)}
}

func (i *instance) Set(name ast.Token, value any) {
	if i.fields == nil {
		i.fields = make(map[string]any)
	}
	i.fields[name.Lexeme] = value
}
