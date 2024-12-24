package ast

import "github.com/araminian/monkey-lang/token"

type Node interface {
	// TokenLiteral returns the literal token of the node
	TokenLiteral() string
}

type Statement interface {
	Node
	// a dummy method to make sure that only statements implement this interface
	statementNode()
}

type Expression interface {
	Node
	// a dummy method to make sure that only expressions implement this interface
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// Identifier is a node for an identifier expression
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// LetStatement is a node for a let statement
type LetStatement struct {
	Token token.Token
	Name  *Identifier // identifier of the binding
	Value Expression  // expression that produces the value of the binding
}

func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}
