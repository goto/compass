package queryexpr

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

type ExprStr interface {
	String() string
	ToQuery() (string, error)
	Validate() error
}

type ExprVisitor struct {
	IdentifiersWithOperator map[string]string // Key: Identifier, Value: Operator
}

type ExprParam map[string]interface{}

func ValidateAndGetQueryFromExpr(exprStr ExprStr) (string, error) {
	if err := exprStr.Validate(); err != nil {
		return "", err
	}
	sqlQuery, err := exprStr.ToQuery()
	if err != nil {
		return "", err
	}

	return sqlQuery, nil
}

// Visit is implementation Visitor interface from expr-lang/expr lib, used by ast.Walk
func (s *ExprVisitor) Visit(node *ast.Node) { //nolint:gocritic
	if n, ok := (*node).(*ast.BinaryNode); ok {
		if left, ok := (n.Left).(*ast.IdentifierNode); ok {
			s.IdentifiersWithOperator[left.Value] = n.Operator
		}
		if right, ok := (n.Right).(*ast.IdentifierNode); ok {
			s.IdentifiersWithOperator[right.Value] = n.Operator
		}
	}
}

func GetIdentifiersMap(queryExpr string) (map[string]string, error) {
	queryExprParsed, err := GetTreeNodeFromQueryExpr(queryExpr)
	if err != nil {
		return nil, err
	}
	queryExprVisitor := &ExprVisitor{IdentifiersWithOperator: make(map[string]string)}
	ast.Walk(&queryExprParsed, queryExprVisitor)
	return queryExprVisitor.IdentifiersWithOperator, nil
}

func GetTreeNodeFromQueryExpr(queryExpr string) (ast.Node, error) {
	parsed, err := parser.Parse(queryExpr)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}

	return parsed.Node, nil
}

func GetQueryExprResult(fn string) (any, error) {
	env := make(ExprParam)
	compile, err := expr.Compile(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to compile function '%s': %w", fn, err)
	}

	result, err := expr.Run(compile, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate function '%s': %w", fn, err)
	}

	return result, nil
}
