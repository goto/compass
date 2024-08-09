package queryexpr

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

type ExprStr interface {
	ToQuery() (string, error)
	Validate() error
}

type QueryExpr struct {
	Identifiers []string
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
func (s *QueryExpr) Visit(node *ast.Node) { //nolint:gocritic
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		s.Identifiers = append(s.Identifiers, n.Value)
	}
}

func GetIdentifiers(queryExpr string) ([]string, error) {
	queryExprParsed, err := GetTreeNodeFromQueryExpr(queryExpr)
	if err != nil {
		return nil, err
	}
	queryExprVisitor := &QueryExpr{}
	ast.Walk(&queryExprParsed, queryExprVisitor)
	return queryExprVisitor.Identifiers, nil
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
