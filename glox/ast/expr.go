package ast

type Expr interface {
	Accept(visitor ExprVisitor) interface{}
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (a AssignExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitAssignExpr(a)
}

type BinaryExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b BinaryExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitBinaryExpr(b)
}

type CallExpr struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
}

func (c CallExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitCallExpr(c)
}

type GetExpr struct {
	Object Expr
	Name   Token
}

func (g GetExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGetExpr(g)
}

type GroupingExpr struct {
	Expression Expr
}

func (g GroupingExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGroupingExpr(g)
}

type LiteralExpr struct {
	Value interface{}
}

func (l LiteralExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLiteralExpr(l)
}

type LogicalExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (l LogicalExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLogicalExpr(l)
}

type SetExpr struct {
	Object Expr
	Name   Token
	Value  Expr
}

func (s SetExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitSetExpr(s)
}

type UnaryExpr struct {
	Operator Token
	Right    Expr
}

func (u UnaryExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitUnaryExpr(u)
}

type VariableExpr struct {
	Name Token
}

func (v VariableExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitVariableExpr(v)
}

type ExprVisitor interface {
	VisitAssignExpr(expr AssignExpr) interface{}
	VisitBinaryExpr(expr BinaryExpr) interface{}
	VisitCallExpr(expr CallExpr) interface{}
	VisitGetExpr(expr GetExpr) interface{}
	VisitGroupingExpr(expr GroupingExpr) interface{}
	VisitLiteralExpr(expr LiteralExpr) interface{}
	VisitLogicalExpr(expr LogicalExpr) interface{}
	VisitSetExpr(expr SetExpr) interface{}
	VisitUnaryExpr(expr UnaryExpr) interface{}
	VisitVariableExpr(expr VariableExpr) interface{}
}
