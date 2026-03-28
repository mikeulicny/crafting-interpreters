package ast

type Stmt interface {
	Accept(visitor StmtVisitor) interface{}
}

type BlockStmt struct {
	Statements []Stmt
}

func (b BlockStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitBlockStmt(b)
}

type ExpressionStmt struct {
	Expr Expr
}

func (e ExpressionStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitExpressionStmt(e)
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

func (f FunctionStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitFunctionStmt(f)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (i IfStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitIfStmt(i)
}

type PrintStmt struct {
	Expr Expr
}

func (p PrintStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitPrintStmt(p)
}

type ReturnStmt struct {
	Keyword Token
	Value   Expr
}

func (r ReturnStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitReturnStmt(r)
}

type VarStmt struct {
	Name        Token
	Initializer Expr
}

func (v VarStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitVarStmt(v)
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (w WhileStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitWhileStmt(w)
}

type StmtVisitor interface {
	VisitBlockStmt(stmt BlockStmt) interface{}
	VisitExpressionStmt(stmt ExpressionStmt) interface{}
	VisitFunctionStmt(stmt FunctionStmt) interface{}
	VisitIfStmt(stmt IfStmt) interface{}
	VisitPrintStmt(stmt PrintStmt) interface{}
	VisitReturnStmt(stmt ReturnStmt) interface{}
	VisitVarStmt(stmt VarStmt) interface{}
	VisitWhileStmt(stmt WhileStmt) interface{}
}
