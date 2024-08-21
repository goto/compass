package queryexpr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

var (
	errFailedGenerateESQuery = errors.New("failed to generate Elasticsearch query")
	errCannotConvertNilQuery = errors.New("cannot convert nil to query")
)

type ExprStr interface {
	String() string
	ToQuery() (string, error)
	Validate() error
}

type exprVisitor struct {
	identifiersWithOperator map[string]string // Key: Identifier, Value: Operator
}

type exprParam map[string]interface{}

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
func (s *exprVisitor) Visit(node *ast.Node) { //nolint:gocritic
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		if left, ok := (n.Left).(*ast.IdentifierNode); ok {
			s.identifiersWithOperator[left.Value] = n.Operator
		}
		if right, ok := (n.Right).(*ast.IdentifierNode); ok {
			s.identifiersWithOperator[right.Value] = n.Operator
		}
	case *ast.UnaryNode:
		if binaryNode, ok := (n.Node).(*ast.BinaryNode); ok {
			if strings.ToUpper(binaryNode.Operator) == "IN" {
				notInOperator := "NOT IN"
				if left, ok := (binaryNode.Left).(*ast.IdentifierNode); ok {
					s.identifiersWithOperator[left.Value] = notInOperator
				}
				if right, ok := (binaryNode.Right).(*ast.IdentifierNode); ok {
					s.identifiersWithOperator[right.Value] = notInOperator
				}
			}
		}
	}
}

func GetIdentifiersMap(queryExpr string) (map[string]string, error) {
	queryExprParsed, err := getTreeNodeFromQueryExpr(queryExpr)
	if err != nil {
		return nil, err
	}
	queryExprVisitor := &exprVisitor{identifiersWithOperator: make(map[string]string)}
	ast.Walk(&queryExprParsed, queryExprVisitor)
	return queryExprVisitor.identifiersWithOperator, nil
}

func getTreeNodeFromQueryExpr(queryExpr string) (ast.Node, error) {
	parsed, err := parser.Parse(queryExpr)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}

	return parsed.Node, nil
}

// getQueryExprResult used for getting the result of query expr operation.
// The playground can be accessed at https://expr-lang.org/playground
//
// Example:
//
//	queryExprOperation := findLast([1, 2, 3, 4], # > 2)
//	result := getQueryExprResult(queryExprOperation)
//
// Result:
//
//	4
func getQueryExprResult(queryExprOperation string) (any, error) {
	env := make(exprParam)
	compile, err := expr.Compile(queryExprOperation)
	if err != nil {
		return nil, fmt.Errorf("failed to compile function '%s': %w", queryExprOperation, err)
	}

	result, err := expr.Run(compile, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate function '%s': %w", queryExprOperation, err)
	}

	if t, ok := result.(time.Time); ok {
		return t.Format(time.RFC3339), nil
	}

	return result, nil
}
