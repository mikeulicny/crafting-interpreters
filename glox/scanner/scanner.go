package scanner

import (
	"fmt"
	"strconv"

	"glox/ast"
)


type Scanner struct {
	source  string
	tokens  []ast.Token
	start   int
	current int
	line    int
}

func New(source string) Scanner {
	return Scanner{source: source, line: 1}
}

func (s *Scanner) ScanTokens() []ast.Token {
	for !s.isAtEnd() {
		// Beginning of next lexeme
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, ast.Token{TokenType: ast.EOF, Line: s.line})
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {
	char := s.advance()
	switch char {
	case '(':
		s.addToken(ast.LEFT_PAREN)
	case ')':
		s.addToken(ast.RIGHT_PAREN)
	case '{':
		s.addToken(ast.LEFT_BRACE)
	case '}':
		s.addToken(ast.RIGHT_BRACE)
	case ',':
		s.addToken(ast.COMMA)
	case '.':
		s.addToken(ast.DOT)
	case '-':
		s.addToken(ast.MINUS)
	case '+':
		s.addToken(ast.PLUS)
	case ';':
		s.addToken(ast.SEMICOLON)
	case '*':
		s.addToken(ast.STAR)

	// cases with look ahead
	case '!':
		var next ast.TokenType
		if s.match('=') {
			next = ast.BANG_EQUAL
		} else {
			next = ast.BANG
		}
		s.addToken(next)
	case '=':
		var next ast.TokenType
		if s.match('=') {
			next = ast.EQUAL_EQUAL
		} else {
			next = ast.EQUAL
		}
		s.addToken(next)
	case '<':
		var next ast.TokenType
		if s.match('=') {
			next = ast.LESS_EQUAL
		} else {
			next = ast.LESS
		}
		s.addToken(next)
	case '>':
		var next ast.TokenType
		if s.match('=') {
			next = ast.GREATER_EQUAL
		} else {
			next = ast.EQUAL
		}
		s.addToken(next)

	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(ast.SLASH)
		}

	// Whitespace
	case ' ', '\r', '\t':

	case '\n':
		s.line++

	case '"':
		s.string()

	default:
		if s.isDigit(char) {
			s.number()
		} else if s.isAlpha(char) {
			s.identifier()
		} else {
			fmt.Printf("[Line: %d] Unexpected character", s.line)
		}
	}
}

func (s *Scanner) advance() byte {
	curr := s.source[s.current]
	s.current++
	return curr
}

func (s *Scanner) addToken(token ast.TokenType) {
	s.addTokenWithLiteral(token, nil)
}

func (s *Scanner) addTokenWithLiteral(token ast.TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, ast.Token{
		TokenType: token,
		Lexeme:    string(text),
		Literal:   literal,
		Line:      s.line,
	})
}

func (s *Scanner) match(char byte) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source[s.current] != char {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return '\000'
	}
	return s.source[s.current+1]
}

func (s *Scanner) string() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		fmt.Printf("[Line: %d] Non-terminated string", s.line)
		return
	}

	// The closing "
	s.advance()

	// Trim surrounding quotes
	val := s.source[s.start+1 : s.current-1]
	s.addTokenWithLiteral(ast.STRING, val)
}

func (s *Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) number() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if (s.peek() == '.') && s.isDigit(s.peekNext()) {
		// consume the '.'
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	val, _ := strconv.ParseFloat(s.source[s.start:s.current], 64)
	s.addTokenWithLiteral(ast.NUMBER, val)
}

func (s *Scanner) isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func (s *Scanner) isAlphaNumeric(char byte) bool {
	return s.isAlpha(char) || s.isDigit(char)
}

var keywords = map[string]ast.TokenType{
	"and":    ast.AND,
	"class":  ast.CLASS,
	"else":   ast.ELSE,
	"false":  ast.FALSE,
	"for":    ast.FOR,
	"fun":    ast.FUN,
	"if":     ast.IF,
	"nil":    ast.NIL,
	"or":     ast.OR,
	"print":  ast.PRINT,
	"return": ast.RETURN,
	"super":  ast.SUPER,
	"this":   ast.THIS,
	"true":   ast.TRUE,
	"var":    ast.VAR,
	"while":  ast.WHILE,
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	tokenType, found := keywords[text]
	if !found {
		tokenType = ast.IDENTIFIER
	}

	s.addToken(tokenType)
}

