package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Expression interface {
	Equal(expression Expression) bool
	String() string
	Pos() QueryScannerPos
}

type Operator struct {
	operator   *QueryToken
	precedence uint
}

type Identifier struct {
	identifier *QueryToken
}

type StringLiteral struct {
	value *QueryToken
}

type NumberLiteral struct {
	value  *QueryToken
	number float64
}

type ParenExpression struct {
	expression Expression
}

type BinaryExpression struct {
	operator *Operator
	lhs      Expression
	rhs      Expression
}

func (operator *Operator) Equal(expression Expression) bool {
	other, ok := expression.(*Operator)
	if !ok {
		return false
	}

	return operator.operator.value == other.operator.value &&
		operator.precedence == other.precedence
}

func (operator *Operator) String() string {
	return fmt.Sprintf("%v[%v]", operator.operator.value, operator.precedence)
}

func (operator *Operator) Pos() QueryScannerPos {
	return operator.operator.startPos
}

func (identifier *Identifier) Equal(expression Expression) bool {
	other, ok := expression.(*Identifier)
	if !ok {
		return false
	}

	return identifier.identifier.value == other.identifier.value
}

func (identifier *Identifier) String() string {
	return identifier.identifier.value
}

func (identifier *Identifier) Pos() QueryScannerPos {
	return identifier.identifier.startPos
}

func (numberLiteral *NumberLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*NumberLiteral)
	if !ok {
		return false
	}

	return numberLiteral.value.value == other.value.value &&
		numberLiteral.number == other.number
}

func (numberLiteral *NumberLiteral) String() string {
	return numberLiteral.value.value
}

func (numberLiteral *NumberLiteral) Pos() QueryScannerPos {
	return numberLiteral.value.startPos
}

func (stringLiteral *StringLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*StringLiteral)
	if !ok {
		return false
	}

	return stringLiteral.value.value == other.value.value
}

func (stringLiteral *StringLiteral) String() string {
	return fmt.Sprintf("\"%v\"", strings.Replace(stringLiteral.value.value, "\"", "\\\"", -1))
}

func (stringLiteral *StringLiteral) Pos() QueryScannerPos {
	return stringLiteral.value.startPos
}

func (parenExpression *ParenExpression) Equal(expression Expression) bool {
	other, ok := expression.(*ParenExpression)
	if !ok {
		return false
	}

	return parenExpression.expression.Equal(other.expression)
}

func (parenExpression *ParenExpression) String() string {
	return fmt.Sprintf("(%v)", parenExpression.expression)
}

func (parenExpression *ParenExpression) Pos() QueryScannerPos {
	return parenExpression.expression.Pos()
}

func (binaryExpression *BinaryExpression) Equal(expression Expression) bool {
	other, ok := expression.(*BinaryExpression)
	if !ok {
		return false
	}

	return binaryExpression.operator.Equal(other.operator) &&
		binaryExpression.lhs.Equal(other.lhs) &&
		binaryExpression.rhs.Equal(other.rhs)
}

func (binaryExpression *BinaryExpression) String() string {
	return fmt.Sprintf("(%v %v %v)", binaryExpression.lhs,
		binaryExpression.operator, binaryExpression.rhs)
}

func (binaryExpression *BinaryExpression) Pos() QueryScannerPos {
	return binaryExpression.operator.operator.startPos
}

func (binaryExpression *BinaryExpression) IsComparison() bool {
	return isComparisonOperator(binaryExpression.operator.operator)
}

var operatorPrecedence = map[QueryTokenType]uint{
	QTK_CMP_EQ: 3,
	QTK_CMP_NE: 3,
	QTK_CMP_GT: 3,
	QTK_CMP_GE: 3,
	QTK_CMP_LT: 3,
	QTK_CMP_LE: 3,

	QTK_AND: 2,
	QTK_OR:  1,
}

type QueryParser struct {
	scanner          *QueryScanner
	tokenBuffer      []*QueryToken
	tokenBufferIndex int
}

func NewQueryParser(reader io.Reader) *QueryParser {
	return &QueryParser{
		scanner: NewQueryScanner(reader),
	}
}

func (parser *QueryParser) scan() (token *QueryToken, err error) {
	if parser.tokenBufferIndex < len(parser.tokenBuffer) {
		token = parser.tokenBuffer[parser.tokenBufferIndex]
		parser.tokenBufferIndex++
		return
	}

	for {
		token, err = parser.scanner.Scan()
		if err != nil {
			return
		}

		if token.tokenType != QTK_WHITE_SPACE {
			break
		}
	}

	parser.tokenBuffer = append(parser.tokenBuffer, token)
	parser.tokenBufferIndex++

	return
}

func (parser *QueryParser) unscan() {
	if parser.tokenBufferIndex > 0 {
		parser.tokenBufferIndex--
	}
}

func GenerateQueryError(token *QueryToken, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v:%v: ", token.startPos.line, token.startPos.col))
	buffer.WriteString(fmt.Sprintf(errorMessage, args...))

	if token.err != nil {
		buffer.WriteString(": ")
		buffer.WriteString(token.err.Error())
	}

	return errors.New(buffer.String())
}

func (parser *QueryParser) Parse() (expression Expression, eof bool, err error) {
	token, err := parser.scan()

	switch {
	case err != nil:
	case token.tokenType == QTK_EOF:
		eof = true
	default:
		parser.unscan()
		expression, err = parser.parseExpression()
	}

	return
}

// Based on parsing code in influxdb
// See https://github.com/influxdata/influxdb/blob/master/influxql/parser.go

func (parser *QueryParser) parseExpression() (expression Expression, err error) {
	root := &BinaryExpression{}

	root.rhs, err = parser.parseUnaryExpression()
	if err != nil {
		return
	}

	for {
		var token *QueryToken
		token, err = parser.scan()
		if err != nil {
			return
		} else if !isOperatorToken(token) {
			if token.tokenType == QTK_EOF || token.tokenType == QTK_RPAREN {
				expression = root.rhs

				if token.tokenType == QTK_RPAREN {
					parser.unscan()
				}
			} else {
				err = GenerateQueryError(token, "Expected operator but found: %v", token.Value())
			}

			return
		}

		var operator *Operator
		operator, err = createOperator(token)
		if err != nil {
			return
		}

		var rhs Expression
		rhs, err = parser.parseUnaryExpression()
		if err != nil {
			return
		}

		for node := root; ; {
			nodeRhs, ok := node.rhs.(*BinaryExpression)

			if !ok || nodeRhs.operator.precedence >= operator.precedence {
				node.rhs = &BinaryExpression{
					operator: operator,
					lhs:      node.rhs,
					rhs:      rhs,
				}

				break
			}

			node = nodeRhs
		}
	}
}

func (parser *QueryParser) parseUnaryExpression() (expression Expression, err error) {
	token, err := parser.scan()
	if err != nil {
		return
	}

	if token.tokenType == QTK_LPAREN {
		var parenExpression Expression
		parenExpression, err = parser.parseExpression()
		if err != nil {
			return
		}

		token, err = parser.scan()
		if err != nil {
			return
		}

		if token.tokenType != QTK_RPAREN {
			err = GenerateQueryError(token, "Expected ')' but found: %v", token.Value())
			return
		}

		expression = &ParenExpression{parenExpression}
		return
	}

	switch token.tokenType {
	case QTK_IDENTIFIER:
		expression = &Identifier{token}
		return
	case QTK_NUMBER:
		var number float64
		number, err = strconv.ParseFloat(token.value, 64)
		if err != nil {
			return
		}

		expression = &NumberLiteral{token, number}
		return
	case QTK_STRING:
		expression = &StringLiteral{token}
		return
	}

	err = GenerateQueryError(token, "Expected Identifier, String or Number but found: %v", token.value)
	return
}

func createOperator(token *QueryToken) (*Operator, error) {
	if !isOperatorToken(token) {
		return nil, GenerateQueryError(token, "Expected operator token but found: %v", token.value)
	}

	precedence, ok := operatorPrecedence[token.tokenType]
	if !ok {
		precedence = uint(QTK_OR - 1)
	}

	operator := &Operator{
		operator:   token,
		precedence: precedence,
	}

	return operator, nil
}

func isOperatorToken(token *QueryToken) bool {
	return isComparisonOperator(token) || isLogicalOperator(token)
}

func isComparisonOperator(token *QueryToken) bool {
	switch token.tokenType {
	case QTK_CMP_EQ, QTK_CMP_NE, QTK_CMP_GT, QTK_CMP_GE, QTK_CMP_LT, QTK_CMP_LE:
		return true
	}

	return false
}

func isLogicalOperator(token *QueryToken) bool {
	switch token.tokenType {
	case QTK_AND, QTK_OR:
		return true
	}

	return false
}

func isValueToken(token *QueryToken) bool {
	switch token.tokenType {
	case QTK_NUMBER, QTK_STRING:
		return true
	}

	return false
}
