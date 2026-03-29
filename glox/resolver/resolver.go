package resolver

import (
	"fmt"
	"io"

	"glox/ast"
	"glox/interpreter"
)

type functionType int

const (
	NONE functionType = iota
	FUNCTION
	INITIALIZER
	METHOD
)

type classType int

const (
	CLASS_NONE classType = iota
	CLASS_CLASS
	CLASS_SUBCLASS
)

type scope map[string]bool

func (s *scope) has(key string) (declared bool, defined bool) {
	return false, false
}

// Creates and sets a variable within the scope to true
func (s scope) set(name string) {
	s[name] = true
}

type stack []scope

// access the top element in the stack
func (s *stack) peek() scope {
	return (*s)[len(*s)-1]
}

// add an element to the stack
func (s *stack) push(scope scope) {
	*s = append(*s, scope)
}

// remove the top element from the stack
func (s *stack) pop() {
	*s = (*s)[:len(*s)-1]
}

// return true if the stack is empty
func (s *stack) isEmpty() bool {
	return len(*s) == 0
}

type Resolver struct {
	interpreter     *interpreter.Interpreter
	scopes          stack
	currentFunction functionType
	currentClass    classType
	stdErr          io.Writer
	hadError        bool
}

func New(interpreter *interpreter.Interpreter) *Resolver {
	return &Resolver{interpreter: interpreter}
}

// Add the variable to the innermost scope, shadowing any outer ones.
// Marked as "not ready yet".
func (r *Resolver) declare(name ast.Token) {
	if r.scopes.isEmpty() {
		return
	}
	scope := r.scopes.peek()
	if _, exists := scope[name.Lexeme]; exists {
		r.error(name, "Already a variable with this name in this scope.")
	}
	scope[name.Lexeme] = false
}

// Set the variable to "ready".
func (r *Resolver) define(name ast.Token) {
	if r.scopes.isEmpty() {
		return
	}
	r.scopes.peek()[name.Lexeme] = true
}

func (r *Resolver) VisitBlockStmt(stmt ast.BlockStmt) interface{} {
	r.beginScope()
	r.ResolveStatements(stmt.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) VisitClassStmt(stmt ast.ClassStmt) interface{} {
	enclosingClass := r.currentClass
	r.currentClass = CLASS_CLASS

	r.declare(stmt.Name)
	r.define(stmt.Name)

	if (stmt.Superclass != nil) && (stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme) {
		r.error(stmt.Superclass.Name, "A class can't inherit from itself.")
	}

	if stmt.Superclass != nil {
		r.currentClass = CLASS_SUBCLASS
		r.resolveExpr(stmt.Superclass)
	}

	if stmt.Superclass != nil {
		r.beginScope()
		defer func() { r.endScope() }()
		r.scopes.peek().set("super")
	}

	r.beginScope()
	r.scopes.peek().set("this")

	fnType := METHOD
	for _, method := range stmt.Methods {
		if method.Name.Lexeme == "init" {
			fnType = INITIALIZER
		}
		r.resolveFunction(method, fnType)
	}
	r.currentClass = enclosingClass
	r.endScope()
	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt ast.ExpressionStmt) interface{} {
	r.resolveExpr(stmt.Expr)
	return nil
}

// Walk a list of statements and resolve each one
func (r *Resolver) ResolveStatements(statements []ast.Stmt) (hadError bool) {
	for _, stmt := range statements {
		r.resolveStmt(stmt)
	}
	return r.hadError
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	expr.Accept(r)
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) resolveLocal(expr ast.Expr, name ast.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		s := r.scopes[i]
		if _, exists := s[name.Lexeme]; exists {
			depth := len(r.scopes) - 1 - i
			r.interpreter.Resolve(expr, depth)
			return
		}
	}
}

func (r *Resolver) resolveFunction(function ast.FunctionStmt, fnType functionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = fnType
	defer func() { r.currentFunction = enclosingFunction }()

	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}
	r.ResolveStatements(function.Body)
	r.endScope()
}

// Create a new scope on the stack
func (r *Resolver) beginScope() {
	r.scopes.push(make(scope))
}

// Removes the current scope from the stack
func (r *Resolver) endScope() {
	r.scopes.pop()
}

func (r *Resolver) VisitFunctionStmt(stmt ast.FunctionStmt) interface{} {
	r.declare(stmt.Name)
	r.define(stmt.Name) // eagerly call define to allow recursion
	r.resolveFunction(stmt, FUNCTION)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt ast.IfStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt(stmt.ElseBranch)
	}
	return nil
}

func (r *Resolver) VisitPrintStmt(stmt ast.PrintStmt) interface{} {
	r.resolveExpr(stmt.Expr)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt ast.ReturnStmt) interface{} {
	if r.currentFunction == NONE {
		r.error(stmt.Keyword, "Can't return from top-level code.")
	}
	if stmt.Value != nil {
		if r.currentFunction == INITIALIZER {
			r.error(stmt.Keyword, "Can't return a value from an initializer.")
		}
		r.resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) VisitVarStmt(stmt ast.VarStmt) interface{} {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt ast.WhileStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	return nil
}

func (r *Resolver) VisitAssignExpr(expr ast.AssignExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitBinaryExpr(expr ast.BinaryExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitCallExpr(expr ast.CallExpr) interface{} {
	r.resolveExpr(expr.Callee)
	for _, arg := range expr.Arguments {
		r.resolveExpr(arg)
	}
	return nil
}

func (r *Resolver) VisitGetExpr(expr ast.GetExpr) interface{} {
	r.resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitGroupingExpr(expr ast.GroupingExpr) interface{} {
	r.resolveExpr(expr.Expression)
	return nil
}

func (r *Resolver) VisitLiteralExpr(expr ast.LiteralExpr) interface{} {
	// Literals don't mention any variables and contain no subexpressions, so there is nothing to do
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr ast.LogicalExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitSetExpr(expr ast.SetExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitSuperExpr(expr ast.SuperExpr) interface{} {
	if r.currentClass == CLASS_NONE {
		r.error(expr.Keyword, "Can't use 'super' outside of a class.")
	} else if r.currentClass != CLASS_SUBCLASS {
		r.error(expr.Keyword, "Can't use 'super' in a class with no superclass.")
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitThisExpr(expr ast.ThisExpr) interface{} {
	if r.currentClass == CLASS_NONE {
		r.error(expr.Keyword, "Can't use 'this' outside of a class.")
		return nil
	}
	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitUnaryExpr(expr ast.UnaryExpr) interface{} {
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitVariableExpr(expr ast.VariableExpr) interface{} {
	if !r.scopes.isEmpty() {
		if val, exists := r.scopes.peek()[expr.Name.Lexeme]; exists && !val {
			r.error(expr.Name, "Can't read local variable in its own initializer.")
		}
	}
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) error(token ast.Token, message string) {
	var pos string
	if token.TokenType == ast.EOF {
		pos = " at end"
	} else {
		pos = "at '" + token.Lexeme + "'"
	}

	r.stdErr.Write([]byte(fmt.Sprintf("[line %d] Error%s: %s\n", token.Line, pos, message)))
	r.hadError = true
}
