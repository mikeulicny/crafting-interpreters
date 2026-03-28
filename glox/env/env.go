package env

import (
	"errors"
	"glox/ast"
)

var ErrUndefined = errors.New("undefined variable")

type Environment struct {
	Enclosing *Environment
	values    map[string]any
}

func New(enclosing *Environment) *Environment {
	return &Environment{
		Enclosing: enclosing,
		values:    make(map[string]any),
	}
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
}

func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for i := 0; i < distance; i++ {
		env = env.Enclosing
	}
	return env
}

func (e *Environment) Get(name ast.Token) (any, error) {
	if val, ok := e.values[name.Lexeme]; ok {
		return val, nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}
	return nil, ErrUndefined
}

func (e *Environment) GetAt(distance int, name string) interface{} {
	return e.ancestor(distance).values[name]
}

func (e *Environment) Assign(name ast.Token, value any) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, value)
	}
	return ErrUndefined
}

func (e *Environment) AssignAt(distance int, name ast.Token, value any) {
	e.ancestor(distance).values[name.Lexeme] = value
}
