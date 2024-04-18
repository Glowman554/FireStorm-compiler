package parser

import "flc/firestorm/lexer"

type DataType int

const (
	INVALID          = -1
	INT     DataType = iota
	STR
	VOID
	CHR
	PTR
    INT_32
    INT_16
)

func GetDatatypeFromString(t string) DataType {
	switch t {
	case "int":
		return INT
	case "str":
		return STR
	case "void":
		return VOID
	case "chr":
		return CHR
	case "ptr":
		return PTR
    case "i32":
        return INT_32
    case "i16":
        return INT_16
	default:
		panic("Invalid datatype")
	}
}

func IsDatatypeString(t string) bool {
	return t == "int" || t == "str" || t == "void" || t == "chr" || t == "ptr" || t == "i32" || t == "i16"
}

func GetTokenFromDatatype(d DataType) lexer.TokenType {
	switch d {
	case INT:
		return lexer.NUMBER
	case STR:
		return lexer.STRING
	default:
		panic("Invalid")
	}
}

type UnnamedDatatype struct {
	Type    DataType
	IsArray bool
}

type NamedDatatype struct {
	UnnamedDatatype
	Name string
}
