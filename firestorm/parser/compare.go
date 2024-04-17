package parser

import "flc/firestorm/lexer"

type Compare int

const (
	More Compare = iota
	Less
	MoreEquals
	LessEquals
	Equals
	NotEquals
)

func TokenTypeToCompare(t lexer.TokenType) Compare {
	switch t {
	case lexer.MORE:
		return More
	case lexer.LESS:
		return Less
	case lexer.MORE_EQUALS:
		return MoreEquals
	case lexer.LESS_EQUALS:
		return LessEquals
	case lexer.EQUALS:
		return Equals
	case lexer.NOT_EQUALS:
		return NotEquals
	default:
		panic("Invalid compare")
	}
}
