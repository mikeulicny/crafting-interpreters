package parser

import (
	"fmt"

	"glox/ast"
)

/*
 * Parser Grammer
 *
 * program      → declaration* EOF ;
 * declaration  → classDecl | funDecl | varDecl | statement ;
 * classDecl    → "class" IDENTIFIER "{" function* "}" ;
 * funDecl      → "fun" function ;
 * function     → IDENTIFIER "(" parameters? ")" block ;
 * parameters   → IDENTIFIER ( "," IDENTIFIER )* ;
 * varDecl      → "var" IDENTIFIER ( "=" expression )? ";" ;
 * statement    → exprStmt | forStmt | ifStmt | printStmt | returnStmt | whileStmt | block ;
 * forStmt      → "for" "(" ( varDecl | exprStmt | ";" )
 *              | expression? ";"
 *              | expression? ")" statement ;
 * whileStmt    → "while" "(" expression ")" statement ;
 * block        → "{" declaration* "}" ;
 * exprStmt     → expression ";" ;
 * ifStmt       → "if" "(" expression ")" statement
 *              ( "else" statement )? ;
 * printStmt    → "print" expression ";" ;
 * returnStmt   → "return" expression? ";" ;
 * expression   → assignment ;
 * assignment   → ( call "." )? IDENTIFIER "=" assignment | logic_or ;
 * logic_or     → logic_and ( "or" logic_and )* ;
 * logic_and    → equality ( "and" equality )* ;
 * ternary      → logic_or ( "?" ternary ":" ternary )* ;
 * equality     → comparison ( ( "!=" | "==" ) comparison )* ;
 * comparison   → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
 * term         → factor ( ( "-" | "+" ) factor )* ;
 * factor       → unary ( ( "/" | "*" ) unary )* ;
 * unary        → ( "!" | "-" ) unary | call ;
 * call         → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
 * arguments    → expression ( "," expression )* ;
 * primary      → "true" | "false" | "nil" | NUMBER | STRING | "(" expression ")" | IDENTIFIER ;
 */

type parseError struct {
	msg string
}

func (p parseError) Error() string {
	return p.msg
}

type Parser struct {
	tokens   []ast.Token
	current  int
	hadError bool
}

type ParseError struct {
	msg string
}

func (p ParseError) Error() string {
	return p.msg
}

func New(tkns []ast.Token) Parser {
	return Parser{tokens: tkns}
}

func (p *Parser) Parse() ([]ast.Stmt, bool) {
	var statements []ast.Stmt
	for !p.isAtEnd() {
		decl := p.declaration()
		statements = append(statements, decl)
	}
	return statements, p.hadError
}

func (p *Parser) expression() ast.Expr {
	return p.assignment()
}

func (p *Parser) declaration() ast.Stmt {
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(parseError); ok {
				p.hadError = true
				p.synchronize()
			} else {
				panic(err)
			}
		}
	}()

	if p.match(ast.CLASS) {
		return p.classDeclaration()
	}
	if p.match(ast.FUN) {
		return p.function("function")
	}
	if p.match(ast.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) classDeclaration() ast.Stmt {
	name := p.consume(ast.IDENTIFIER, "Expect class name.")
	p.consume(ast.LEFT_BRACE, "Expect '{' before class body.")
	var methods []ast.FunctionStmt
	for !p.check(ast.RIGHT_BRACE) && !p.isAtEnd() {
		method := p.function("method")
		methods = append(methods, method)
	}
	p.consume(ast.RIGHT_BRACE, "Expect '}' after class body.")
	return ast.ClassStmt{Name: name, Methods: methods}
}

func (p *Parser) statement() ast.Stmt {
	if p.match(ast.FOR) {
		return p.forStatement()
	}
	if p.match(ast.IF) {
		return p.ifStatement()
	}
	if p.match(ast.PRINT) {
		return p.printStatement()
	}
	if p.match(ast.RETURN) {
		return p.returnStatement()
	}
	if p.match(ast.WHILE) {
		return p.whileStatement()
	}
	if p.match(ast.LEFT_BRACE) {
		return ast.BlockStmt{Statements: p.block()}
	}
	return p.expressionStatement()
}

func (p *Parser) forStatement() ast.Stmt {
	p.consume(ast.LEFT_PAREN, "Expect '(' after 'for'.")

	var initializer ast.Stmt
	if p.match(ast.SEMICOLON) {
		initializer = nil
	} else if p.match(ast.VAR) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition ast.Expr
	if !p.check(ast.SEMICOLON) {
		condition = p.expression()
	}
	p.consume(ast.SEMICOLON, "Expect ';' after loop condition.")

	var increment ast.Expr
	if !p.check(ast.RIGHT_PAREN) {
		increment = p.expression()
	}
	p.consume(ast.RIGHT_PAREN, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = ast.BlockStmt{Statements: []ast.Stmt{body, ast.ExpressionStmt{Expr: increment}}}
	}

	if condition == nil {
		condition = ast.LiteralExpr{Value: true}
	}
	body = ast.WhileStmt{Condition: condition, Body: body}

	if initializer != nil {
		body = ast.BlockStmt{Statements: []ast.Stmt{initializer, body}}
	}

	return body
}

func (p *Parser) ifStatement() ast.Stmt {
	p.consume(ast.LEFT_PAREN, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(ast.RIGHT_PAREN, "Expect ')' after if condition.")

	thenBranch := p.statement()
	var elseBranch ast.Stmt
	if p.match(ast.ELSE) {
		elseBranch = p.statement()
	}
	return ast.IfStmt{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (p *Parser) whileStatement() ast.Stmt {
	p.consume(ast.LEFT_PAREN, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(ast.RIGHT_PAREN, "Expect ')' after condition.")
	body := p.statement()

	return ast.WhileStmt{Condition: condition, Body: body}
}

func (p *Parser) printStatement() ast.Stmt {
	expr := p.expression()
	p.consume(ast.SEMICOLON, "Expect ';' after value.")
	return ast.PrintStmt{Expr: expr}
}

func (p *Parser) returnStatement() ast.Stmt {
	keyword := p.previous()
	var value ast.Expr
	if !p.check(ast.SEMICOLON) {
		value = p.expression()
	}
	p.consume(ast.SEMICOLON, "Expect ';' after return value.")
	return ast.ReturnStmt{Keyword: keyword, Value: value}
}

func (p *Parser) varDeclaration() ast.Stmt {
	name := p.consume(ast.IDENTIFIER, "Expect variable name.")

	var initializer ast.Expr
	if p.match(ast.EQUAL) {
		initializer = p.expression()
	}

	p.consume(ast.SEMICOLON, "Expect ';' after variable declaration.")
	return ast.VarStmt{Name: name, Initializer: initializer}
}

func (p *Parser) expressionStatement() ast.Stmt {
	expr := p.expression()
	p.consume(ast.SEMICOLON, "Expect ';' after expression.")
	return ast.ExpressionStmt{Expr: expr}
}

func (p *Parser) function(kind string) ast.FunctionStmt {
	name := p.consume(ast.IDENTIFIER, "Expect "+kind+" name.")
	p.consume(ast.LEFT_PAREN, "Expect '(' after "+kind+" name.")
	var params []ast.Token
	if !p.check(ast.RIGHT_PAREN) {
		for {
			if len(params) >= 255 {
				p.error(p.peek(), "Can't have more than 255 parameters.")
			}

			expr := p.consume(ast.IDENTIFIER, "Expect parameter name.")
			params = append(params, expr)

			if !p.match(ast.COMMA) {
				break
			}
		}
	}
	p.consume(ast.RIGHT_PAREN, "Expect ')' after parameters.")

	p.consume(ast.LEFT_BRACE, "Expect '{' before "+kind+" body.")
	body := p.block()
	return ast.FunctionStmt{Name: name, Params: params, Body: body}
}

func (p *Parser) block() []ast.Stmt {
	var statements []ast.Stmt
	for !p.check(ast.RIGHT_BRACE) && !p.isAtEnd() {
		statements = append(statements, p.declaration())
	}
	p.consume(ast.RIGHT_BRACE, "Expect '}' after block.")
	return statements
}

func (p *Parser) assignment() ast.Expr {
	expr := p.or()

	if p.match(ast.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		if varExpr, ok := expr.(ast.VariableExpr); ok {
			return ast.AssignExpr{Name: varExpr.Name, Value: value}
		} else if getExpr, ok := expr.(ast.GetExpr); ok {
			return ast.SetExpr{
				Object: getExpr.Object,
				Name:   getExpr.Name,
				Value:  value,
			}
		}

		p.error(equals, "Invalid assignment target.")
	}

	return expr
}

func (p *Parser) or() ast.Expr {
	expr := p.and()

	for p.match(ast.OR) {
		operator := p.previous()
		right := p.equality()
		expr = ast.LogicalExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) and() ast.Expr {
	expr := p.equality()

	for p.match(ast.AND) {
		operator := p.previous()
		right := p.equality()
		expr = ast.LogicalExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparison()

	for p.match(ast.BANG_EQUAL, ast.EQUAL_EQUAL) {
		op := p.previous()
		right := p.comparison()
		expr = ast.BinaryExpr{Left: expr, Operator: op, Right: right}
	}

	return expr
}

func (p *Parser) comparison() ast.Expr {
	expr := p.term()

	for p.match(ast.GREATER, ast.GREATER_EQUAL, ast.LESS, ast.LESS_EQUAL) {
		op := p.previous()
		right := p.term()
		expr = ast.BinaryExpr{Left: expr, Operator: op, Right: right}
	}

	return expr
}

func (p *Parser) term() ast.Expr {
	expr := p.factor()

	for p.match(ast.MINUS, ast.PLUS) {
		op := p.previous()
		right := p.factor()
		expr = ast.BinaryExpr{Left: expr, Operator: op, Right: right}
	}

	return expr
}

func (p *Parser) factor() ast.Expr {
	expr := p.unary()

	for p.match(ast.SLASH, ast.STAR) {
		op := p.previous()
		right := p.unary()
		expr = ast.BinaryExpr{Left: expr, Operator: op, Right: right}
	}

	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(ast.BANG, ast.MINUS) {
		op := p.previous()
		right := p.unary()
		return ast.UnaryExpr{Operator: op, Right: right}
	}
	return p.call()
}

func (p *Parser) call() ast.Expr {
	expr := p.primary()

	for {
		if p.match(ast.LEFT_PAREN) {
			expr = p.finishCall(expr)
		} else if p.match(ast.DOT) {
			name := p.consume(ast.IDENTIFIER, "Expect property name after '.'.")
			expr = ast.GetExpr{Object: expr, Name: name}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee ast.Expr) ast.Expr {
	var arguments []ast.Expr

	if !p.check(ast.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				p.error(p.peek(), "Can't have more than 255 arguments.")
			}
			expr := p.expression()
			arguments = append(arguments, expr)
			if !p.match(ast.COMMA) {
				break
			}
		}
	}

	paren := p.consume(ast.RIGHT_PAREN, "Expect ')' after arguments.")

	return ast.CallExpr{Callee: callee, Paren: paren, Arguments: arguments}
}

func (p *Parser) primary() ast.Expr {
	if p.match(ast.FALSE) {
		return ast.LiteralExpr{Value: false}
	}
	if p.match(ast.TRUE) {
		return ast.LiteralExpr{Value: true}
	}
	if p.match(ast.NIL) {
		return ast.LiteralExpr{Value: nil}
	}

	if p.match(ast.NUMBER, ast.STRING) {
		return ast.LiteralExpr{Value: p.previous().Literal}
	}

	if p.match(ast.IDENTIFIER) {
		return ast.VariableExpr{Name: p.previous()}
	}

	if p.match(ast.LEFT_PAREN) {
		expr := p.expression()
		p.consume(ast.RIGHT_PAREN, "Expect ')' after expression")
		return ast.GroupingExpr{Expression: expr}
	}

	p.error(p.peek(), "Expect expression.")
	return nil
}

// Attempt to consume the provided token type.
// If the current token is equal, consume it and advance to the next token.
// If it is not equal, throw an error with a message.
func (p *Parser) consume(tokentype ast.TokenType, message string) ast.Token {
	if p.check(tokentype) {
		return p.advance()
	}
	p.error(p.peek(), message)
	return ast.Token{}
}

// Perform a check() and advance() on all provided token types. If no matches
// are found returns false.
func (p *Parser) match(types ...ast.TokenType) bool {
	for _, tokentype := range types {
		if p.check(tokentype) {
			p.advance()
			return true
		}
	}
	return false
}

// Check if the current token is equivalent to the passed in token.
func (p *Parser) check(t ast.TokenType) bool {
	if p.isAtEnd() {
		return false
	} else {
		return p.peek().TokenType == t
	}
}

// Advance the parser to the next token, and return the current token.
func (p *Parser) advance() ast.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// Check if the current token is EOF
func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == ast.EOF
}

// View the current token
func (p *Parser) peek() ast.Token {
	return p.tokens[p.current]
}

// Get the previous token
func (p *Parser) previous() ast.Token {
	return p.tokens[p.current-1]
}

// Throw a parsing error with a message, identifying the line the error occured on
func (p *Parser) error(token ast.Token, message string) {
	var pos string
	if token.TokenType == ast.EOF {
		pos = " at end"
	} else {
		pos = " at '" + token.Lexeme + "'"
	}

	err := ParseError{msg: fmt.Sprintf("[line %d] Error%s: %s\n", token.Line+1, pos, message)}
	panic(err)
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().TokenType == ast.SEMICOLON {
			return
		}

		switch p.peek().TokenType {
		case ast.CLASS, ast.FUN, ast.VAR, ast.FOR, ast.IF, ast.WHILE, ast.PRINT, ast.RETURN:
			return
		}

		p.advance()
	}
}
