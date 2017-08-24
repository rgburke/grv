package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Expression represents a parsed query expression
type Expression interface {
	Equal(expression Expression) bool
	String() string
	Pos() QueryScannerPos
}

// LogicalOperatorExpression is an expression which resolves to a boolean value
type LogicalOperatorExpression interface {
	RHS() Expression
	OperatorPrecedence() uint
	SetRHS(expression Expression)
}

// Operator represents a unary or binary operator
type Operator struct {
	operator   *QueryToken
	precedence uint
}

// Identifier represents a field name
type Identifier struct {
	identifier *QueryToken
}

// StringLiteral is a literal string value
type StringLiteral struct {
	value *QueryToken
}

// NumberLiteral is a numeric value
type NumberLiteral struct {
	value  *QueryToken
	number float64
}

// ParenExpression represents an expression contained in parentheses
type ParenExpression struct {
	expression Expression
}

// UnaryExpression groups an operator and a single operand
type UnaryExpression struct {
	operator   *Operator
	expression Expression
}

// BinaryExpression groups an operator and two operands
type BinaryExpression struct {
	operator *Operator
	lhs      Expression
	rhs      Expression
}

// Equal returns true if this expression is equal to the provided expression
func (operator *Operator) Equal(expression Expression) bool {
	other, ok := expression.(*Operator)
	if !ok {
		return false
	}

	return operator.operator.value == other.operator.value &&
		operator.precedence == other.precedence
}

// String returns the string representation of the operator with precedence
func (operator *Operator) String() string {
	return fmt.Sprintf("%v[%v]", operator.operator.value, operator.precedence)
}

// Pos returns the position this token appeared at in the input stream
func (operator *Operator) Pos() QueryScannerPos {
	return operator.operator.startPos
}

// Equal returns true if this expression is equal to the provided expression
func (identifier *Identifier) Equal(expression Expression) bool {
	other, ok := expression.(*Identifier)
	if !ok {
		return false
	}

	return identifier.identifier.value == other.identifier.value
}

// String returns the field name
func (identifier *Identifier) String() string {
	return identifier.identifier.value
}

// Pos returns the position this token appeared at in the input stream
func (identifier *Identifier) Pos() QueryScannerPos {
	return identifier.identifier.startPos
}

// Equal returns true if this expression is equal to the provided expression
func (numberLiteral *NumberLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*NumberLiteral)
	if !ok {
		return false
	}

	return numberLiteral.value.value == other.value.value &&
		numberLiteral.number == other.number
}

// String returns the string representation of the numberic value stored
func (numberLiteral *NumberLiteral) String() string {
	return numberLiteral.value.value
}

// Pos returns the position this token appeared at in the input stream
func (numberLiteral *NumberLiteral) Pos() QueryScannerPos {
	return numberLiteral.value.startPos
}

// Equal returns true if this expression is equal to the provided expression
func (stringLiteral *StringLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*StringLiteral)
	if !ok {
		return false
	}

	return stringLiteral.value.value == other.value.value
}

// String returns the processed string value
func (stringLiteral *StringLiteral) String() string {
	return fmt.Sprintf("\"%v\"", strings.Replace(stringLiteral.value.value, "\"", "\\\"", -1))
}

// Pos returns the position this token appeared at in the input stream
func (stringLiteral *StringLiteral) Pos() QueryScannerPos {
	return stringLiteral.value.startPos
}

// Equal returns true if this expression is equal to the provided expression
func (parenExpression *ParenExpression) Equal(expression Expression) bool {
	other, ok := expression.(*ParenExpression)
	if !ok {
		return false
	}

	return parenExpression.expression.Equal(other.expression)
}

// String returns the expression in parenthesis
func (parenExpression *ParenExpression) String() string {
	return fmt.Sprintf("(%v)", parenExpression.expression)
}

// Pos returns the position the child expression appeared at in the input stream
func (parenExpression *ParenExpression) Pos() QueryScannerPos {
	return parenExpression.expression.Pos()
}

// Equal returns true if this expression is equal to the provided expression
func (unaryExpression *UnaryExpression) Equal(expression Expression) bool {
	other, ok := expression.(*UnaryExpression)
	if !ok {
		return false
	}

	return unaryExpression.operator.Equal(other.operator) &&
		unaryExpression.expression.Equal(other.expression)
}

// String returns the unary expression in parenthesis
func (unaryExpression *UnaryExpression) String() string {
	return fmt.Sprintf("(%v %v)", unaryExpression.operator, unaryExpression.expression)
}

// Pos returns the position this expressions operator appeared in the input stream
func (unaryExpression *UnaryExpression) Pos() QueryScannerPos {
	return unaryExpression.operator.operator.startPos
}

// RHS returns the operand for the unary expression
func (unaryExpression *UnaryExpression) RHS() Expression {
	return unaryExpression.expression
}

// OperatorPrecedence returns the operator precedence of this expressions operator
func (unaryExpression *UnaryExpression) OperatorPrecedence() uint {
	return unaryExpression.operator.precedence
}

// SetRHS sets the operand of this expression
func (unaryExpression *UnaryExpression) SetRHS(expression Expression) {
	unaryExpression.expression = expression
}

// Equal returns true if this expression is equal to the provided expression
func (binaryExpression *BinaryExpression) Equal(expression Expression) bool {
	other, ok := expression.(*BinaryExpression)
	if !ok {
		return false
	}

	return binaryExpression.operator.Equal(other.operator) &&
		binaryExpression.lhs.Equal(other.lhs) &&
		binaryExpression.rhs.Equal(other.rhs)
}

// String returns the binary expression in parenthesis
func (binaryExpression *BinaryExpression) String() string {
	return fmt.Sprintf("(%v %v %v)", binaryExpression.lhs,
		binaryExpression.operator, binaryExpression.rhs)
}

// Pos returns the position this token appeared at in the input stream
func (binaryExpression *BinaryExpression) Pos() QueryScannerPos {
	return binaryExpression.operator.operator.startPos
}

// RHS returns the right hand side operand
func (binaryExpression *BinaryExpression) RHS() Expression {
	return binaryExpression.rhs
}

// OperatorPrecedence returns the precedence of this expressions operator
func (binaryExpression *BinaryExpression) OperatorPrecedence() uint {
	return binaryExpression.operator.precedence
}

// SetRHS sets the right hand side operand
func (binaryExpression *BinaryExpression) SetRHS(expression Expression) {
	binaryExpression.rhs = expression
}

// IsComparison returns true if the operator for this expression is a comparison operator
func (binaryExpression *BinaryExpression) IsComparison() bool {
	return isComparisonOperator(binaryExpression.operator.operator)
}

var operatorPrecedence = map[QueryTokenType]uint{
	QtkCmpEq: 4,
	QtkCmpNe: 4,
	QtkCmpGt: 4,
	QtkCmpGe: 4,
	QtkCmpLt: 4,
	QtkCmpLe: 4,

	QtkCmpGlob:   4,
	QtkCmpRegexp: 4,

	QtkNot: 3,

	QtkAnd: 2,
	QtkOr:  1,
}

// QueryParser parses an input stream into a sequence of expressions
type QueryParser struct {
	scanner          *QueryScanner
	tokenBuffer      []*QueryToken
	tokenBufferIndex int
}

// NewQueryParser creates a new instance
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

		if token.tokenType != QtkWhiteSpace {
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

func generateQueryError(token *QueryToken, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v:%v: ", token.startPos.line, token.startPos.col))
	buffer.WriteString(fmt.Sprintf(errorMessage, args...))

	if token.err != nil {
		buffer.WriteString(": ")
		buffer.WriteString(token.err.Error())
	}

	return errors.New(buffer.String())
}

// Parse parses and returns the next expression from the input stream or sets eof = true is the end of the input steam is reached
func (parser *QueryParser) Parse() (expression Expression, eof bool, err error) {
	token, err := parser.scan()

	switch {
	case err != nil:
	case token.tokenType == QtkEOF:
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
			if token.tokenType == QtkEOF || token.tokenType == QtkRparen {
				expression = root.rhs

				if token.tokenType == QtkRparen {
					parser.unscan()
				}
			} else {
				err = generateQueryError(token, "Expected operator but found: %v", token.Value())
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

		for node := LogicalOperatorExpression(root); ; {
			nodeRHS, ok := node.RHS().(LogicalOperatorExpression)

			if !ok || nodeRHS.OperatorPrecedence() >= operator.precedence {
				node.SetRHS(&BinaryExpression{
					operator: operator,
					lhs:      node.RHS(),
					rhs:      rhs,
				})

				break
			}

			node = nodeRHS
		}
	}
}

func (parser *QueryParser) parseUnaryExpression() (expression Expression, err error) {
	token, err := parser.scan()
	if err != nil {
		return
	}

	switch token.tokenType {
	case QtkLparen:
		var parenExpression Expression
		parenExpression, err = parser.parseExpression()
		if err != nil {
			return
		}

		token, err = parser.scan()
		if err != nil {
			return
		}

		if token.tokenType != QtkRparen {
			err = generateQueryError(token, "Expected ')' but found: %v", token.Value())
			return
		}

		expression = &ParenExpression{parenExpression}
		return
	case QtkNot:
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

		expression = &UnaryExpression{
			operator:   operator,
			expression: rhs,
		}

		return
	case QtkIdentifier:
		expression = &Identifier{token}
		return
	case QtkNumber:
		var number float64
		number, err = strconv.ParseFloat(token.value, 64)
		if err != nil {
			return
		}

		expression = &NumberLiteral{token, number}
		return
	case QtkString:
		expression = &StringLiteral{token}
		return
	}

	err = generateQueryError(token, "Expected Identifier, String or Number but found: %v", token.value)
	return
}

func createOperator(token *QueryToken) (*Operator, error) {
	if !isOperatorToken(token) {
		return nil, generateQueryError(token, "Expected operator token but found: %v", token.value)
	}

	precedence, ok := operatorPrecedence[token.tokenType]
	if !ok {
		precedence = uint(QtkOr - 1)
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
	case QtkCmpEq, QtkCmpNe, QtkCmpGt, QtkCmpGe, QtkCmpLt, QtkCmpLe,
		QtkCmpGlob, QtkCmpRegexp:
		return true
	}

	return false
}

func isLogicalOperator(token *QueryToken) bool {
	switch token.tokenType {
	case QtkAnd, QtkOr, QtkNot:
		return true
	}

	return false
}
