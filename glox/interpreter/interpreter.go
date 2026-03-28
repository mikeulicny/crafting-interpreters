package interpreter

import (
	"fmt"
	"glox/ast"
	"glox/env"
	"io"
)

type runtimeError struct {
	token ast.Token
	msg   string
}

func (r runtimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", r.msg, r.token.Line)
}

type Return struct {
	Value any
}

type Interpreter struct {
	environment *env.Environment
	globals     *env.Environment
	locals      map[ast.Expr]int
	stdOut      io.Writer
	stdErr      io.Writer
}

func New(stdOut io.Writer, stdErr io.Writer) *Interpreter {
	globals := env.New(nil)
	globals.Define("clock", clock{})

	return &Interpreter{
		environment: globals,
		globals:     globals,
		locals:      make(map[ast.Expr]int),
		stdOut:      stdOut,
		stdErr:      stdErr,
	}
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) (result any, hasRuntimeError bool) {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(runtimeError); ok {
				i.stdErr.Write([]byte(e.Error() + "\n"))
				hasRuntimeError = true
			} else {
				fmt.Printf("Error: %s\n", err)
			}
		}
	}()

	for _, statement := range stmts {
		result = i.execute(statement)
	}

	return
}

// Implement Expr visitor
func (i *Interpreter) VisitAssignExpr(expr ast.AssignExpr) interface{} {
	value := i.evaluate(expr.Value)

	distance, ok := i.locals[expr]; 
	if ok {
		i.environment.AssignAt(distance, expr.Name, value)
	} else {
		if err := i.globals.Assign(expr.Name, value); err != nil {
			panic(err)
		}
	}
	return value
}

func (i *Interpreter) VisitBinaryExpr(expr ast.BinaryExpr) interface{} {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	// Comparison
	case ast.GREATER:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case ast.GREATER_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case ast.LESS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case ast.LESS_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) <= right.(float64)
	case ast.BANG_EQUAL:
		return !i.isEqual(left, right)
	case ast.EQUAL_EQUAL:
		return i.isEqual(left, right)
	// Arithmetic
	case ast.MINUS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case ast.PLUS:
		_, leftFloat := left.(float64)
		_, rightFloat := right.(float64)
		if leftFloat && rightFloat {
			return left.(float64) + right.(float64)
		}
		_, leftString := left.(string)
		_, rightString := right.(string)
		if leftString && rightString {
			return left.(string) + right.(string)
		}
		panic(runtimeError{token: expr.Operator, msg: "Operands must be two numbers or two strings."})
	case ast.SLASH:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)
	case ast.STAR:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)
	}

	// unreachable
	return nil
}

func (i *Interpreter) VisitCallExpr(expr ast.CallExpr) interface{} {
	callee := i.evaluate(expr.Callee)

	arguments := make([]any, len(expr.Arguments))
	for idx, argument := range expr.Arguments {
		arguments[idx] = i.evaluate(argument)
	}


	function, ok := (callee).(callable)
	if !ok {
		panic(runtimeError{token: expr.Paren, msg: "Can only call functions and classes."})
	}
	if len(arguments) != function.arity() {
		panic(runtimeError{ token: expr.Paren, msg: fmt.Sprintf("Expected %d arguments but got %d.", function.arity(), len(arguments)) })
	}
	return function.call(i, arguments)
}

func (i *Interpreter) VisitGroupingExpr(expr ast.GroupingExpr) interface{} {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr ast.LiteralExpr) interface{} {
	return expr.Value
}

func (i *Interpreter) VisitLogicalExpr(expr ast.LogicalExpr) interface{} {
	left := i.evaluate(expr.Left)

	if expr.Operator.TokenType == ast.OR {
		if i.isTruthy(left) {
			return left
		}
	} else {
		if !i.isTruthy(left) {
			return left
		}
	}

	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitUnaryExpr(expr ast.UnaryExpr) interface{} {
	right := i.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	case ast.BANG:
		return !i.isTruthy(right)
	case ast.MINUS:
		return -right.(float64)
	}

	return nil
}

func (i *Interpreter) VisitVariableExpr(expr ast.VariableExpr) interface{} {
	val, err := i.lookupVariable(expr.Name, expr)
	if err != nil {
		panic(err)
	}
	return val
}

func (i *Interpreter) lookupVariable(name ast.Token, expr ast.Expr) (interface{}, error) {
	if distance, ok := i.locals[expr]; ok {
		return i.environment.GetAt(distance, name.Lexeme), nil
	}
	return i.globals.Get(name)
}

func (i *Interpreter) evaluate(expr ast.Expr) interface{} {
	return expr.Accept(i)
}

func (i *Interpreter) execute(stmt ast.Stmt) interface{} {
	return stmt.Accept(i)
}

func (i *Interpreter) Resolve(expr ast.Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) executeBlock(statements []ast.Stmt, env *env.Environment) {
	previous := i.environment
	defer func() {
		i.environment = previous
	}()

	i.environment = env
	for _, statement := range statements {
		i.execute(statement)
	}
}

func (i *Interpreter) VisitBlockStmt(stmt ast.BlockStmt) interface{} {
	i.executeBlock(stmt.Statements, env.New(i.environment))
	return nil
}

func (i *Interpreter) isTruthy(val any) bool {
	if val == nil {
		return false
	}
	if v, ok := val.(bool); ok {
		return v
	}
	return true
}

func (i *Interpreter) isEqual(left any, right any) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil {
		return false
	}
	return left == right
}

func (i *Interpreter) checkNumberOperand(op ast.Token, right any) {
	if _, ok := right.(float64); ok {
		return
	}
	panic(runtimeError{token: op, msg: "Operand must be a number."})
}

func (i *Interpreter) checkNumberOperands(op ast.Token, left, right any) {
	if _, ok := left.(float64); ok {
		if _, ok := right.(float64); ok {
			return
		}
	}
	panic(runtimeError{token: op, msg: "Operands must be a number."})
}

func (i *Interpreter) stringify(val any) string {
	if val == nil {
		return "nil"
	}
	return fmt.Sprint(val)
}

// Implement Stmt visitor
func (i *Interpreter) VisitExpressionStmt(stmt ast.ExpressionStmt) interface{} {
	i.evaluate(stmt.Expr)
	return nil
}

func (i *Interpreter) VisitFunctionStmt(stmt ast.FunctionStmt) interface{} {
	function := function{declaration: stmt, closure: i.environment}
	i.environment.Define(stmt.Name.Lexeme, function)
	return nil
}

func (i *Interpreter) VisitIfStmt(stmt ast.IfStmt) interface{} {
	if i.isTruthy(i.evaluate(stmt.Condition)) {
		i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		i.execute(stmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt ast.PrintStmt) interface{} {
	val := i.evaluate(stmt.Expr)
	i.stdOut.Write([]byte(i.stringify(val) + "\n"))
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt ast.ReturnStmt) interface{} {
	var value any
	if stmt.Value != nil {
		value = i.evaluate(stmt.Value)
	}
	panic(Return{Value: value})
}

func (i *Interpreter) VisitVarStmt(stmt ast.VarStmt) interface{} {
	var value any
	if stmt.Initializer != nil {
		value = i.evaluate(stmt.Initializer)
	}

	i.environment.Define(stmt.Name.Lexeme, value)
	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt ast.WhileStmt) interface{} {
	for i.isTruthy(i.evaluate(stmt.Condition)) {
		i.execute(stmt.Body)
	}
	return nil
}
