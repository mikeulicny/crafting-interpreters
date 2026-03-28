package main

import (
	"bufio"
	"fmt"
	"glox/ast"
	"glox/interpreter"
	"glox/parser"
	"glox/resolver"
	"glox/scanner"
	"go/token"
	"os"
)

type Lox struct {
	interpreter     *interpreter.Interpreter
	hadError        bool
	hadRuntimeError bool
}

func (l *Lox) runFile(path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	l.run(bytes)
}

func (l *Lox) runPrompt() {
	input := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		bytes, err := input.ReadBytes('\n')
		if err != nil {
			break
		}
		l.run(bytes)
	}
}

func main() {
	lox := Lox{
		interpreter: interpreter.New(os.Stdout, os.Stderr),
	}

	args := os.Args[1:]

	if len(args) == 1 {
		lox.runFile(args[0])
	} else {
		lox.runPrompt()
	}
}

func errorHandler(pos token.Position, msg string) {
	fmt.Printf("[ERROR] Syntax error at line: %d - col: %d - %s\n", pos.Line, pos.Column, msg)
}

func (l *Lox) run(source []byte) {
	s := scanner.New(string(source))
	tokens := s.ScanTokens()
	parser := parser.New(tokens)
	var statements []ast.Stmt
	statements, l.hadError = parser.Parse()

	if l.hadError {
		os.Exit(65)
	}

	resolver := resolver.New(l.interpreter)
	l.hadError = resolver.ResolveStatements(statements)

	if l.hadError {
		os.Exit(1)
	}

	_, l.hadRuntimeError = l.interpreter.Interpret(statements)
	if l.hadRuntimeError {
		os.Exit(70)
	}

}
