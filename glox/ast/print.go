package ast

import (
	"fmt"
)

type AstPrinter struct {}

func (p AstPrinter) Print(expr Expr) string {
	return expr.Accept(p).(string)
}

func (p AstPrinter) VisitAssignExpr(expr AssignExpr) interface{} {
	return p.parenthesize("= "+expr.Name.Lexeme, expr.Value)
}

func (p AstPrinter) VisitBinaryExpr(expr BinaryExpr) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) VisitCallExpr(expr CallExpr) interface{} {
	panic("TODO: IMPLEMENT THIS")
}

func (p AstPrinter) VisitGetExpr(expr GetExpr) interface{} {
	panic("TODO: IMPLEMENT THIS")
}

func (p AstPrinter) VisitGroupingExpr(expr GroupingExpr) interface{} {
	return p.parenthesize("group", expr.Expression)
}

func (p AstPrinter) VisitLiteralExpr(expr LiteralExpr) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprint(expr.Value)
}

func (p AstPrinter) VisitLogicalExpr(expr LogicalExpr) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) VisitSetExpr(expr SetExpr) interface{} {
	panic("TODO: IMPLEMENT THIS")
}

func (p AstPrinter) VisitUnaryExpr(expr UnaryExpr) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p AstPrinter) VisitVariableExpr(expr VariableExpr) interface{} {
	return expr.Name.Lexeme
}

func (p AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var str string

	str += "(" + name
	for _, expr := range exprs {
		str += " " + p.Print(expr)
	}
	str += ")"

	return str
}
