package parser

import (
	"fmt"
	"strconv"

	"github.com/araminian/monkey-lang/ast"
	"github.com/araminian/monkey-lang/lexer"
	"github.com/araminian/monkey-lang/token"
)

const (
	// precedence levels
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

type Parser struct {
	l *lexer.Lexer

	// current token
	curToken token.Token
	// next token
	peekToken token.Token

	// errors
	errors []string

	// prefixParseFns is a map of prefix tokens to their corresponding parsing functions
	prefixParseFns map[token.TokenType]prefixParseFn
	// infixParseFns is a map of infix tokens to their corresponding parsing functions
	infixParseFns map[token.TokenType]infixParseFn
}

// RegisterPrefix registers a prefix parsing function for a given token type
func (p *Parser) RegisterPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// RegisterInfix registers an infix parsing function for a given token type
func (p *Parser) RegisterInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	p.errors = []string{}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Register prefix parsing functions
	p.RegisterPrefix(token.IDENT, p.parseIdentifier)
	p.RegisterPrefix(token.INT, p.parseIntegerLiteral)
	p.RegisterPrefix(token.BANG, p.parsePrefixExpression)
	p.RegisterPrefix(token.MINUS, p.parsePrefixExpression)

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// peekError adds an error message to the parser's error list
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// nextToken advances the parser to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: We are skippling expression

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// curTokenIs checks if the current token is of a given type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the next token is of a given type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of a given type and advances the parser if it is
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// TODO: We are skipping the expression, until we encounter a semicolon

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Expression
type (
	// prefixParseFn is a function that is called when the parser encounters a prefix token
	prefixParseFn func() ast.Expression
	// infixParseFn is a function that is called when the parser encounters an infix token
	infixParseFn func(ast.Expression) ast.Expression
)

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	exp.Right = p.parseExpression(PREFIX)

	return exp
}
