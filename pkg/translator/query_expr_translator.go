package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"log"
	"strconv"
	"strings"
)

type QueryExprTranslator struct {
	QueryExpr string
	SqlQuery  strings.Builder
	EsQuery   map[string]interface{}
}

type ExprParam map[string]interface{}

func (q *QueryExprTranslator) getTreeNodeFromQueryExpr() *ast.Node {
	parsed, err := parser.Parse(q.QueryExpr)
	if err != nil {
		log.Fatalf("Error parsing expression: %v", err)
	}

	return &parsed.Node
}

func (q *QueryExprTranslator) ConvertToSQL() (string, error) {
	q.SqlQuery = strings.Builder{}
	q.TranslateToSQL(q.getTreeNodeFromQueryExpr(), q)
	return q.SqlQuery.String(), nil
}

// TranslateToSQL The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (q *QueryExprTranslator) TranslateToSQL(node *ast.Node, translator *QueryExprTranslator) {
	if *node == nil {
		return
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		translator.SqlQuery.WriteString("(")
		q.TranslateToSQL(&n.Left, translator)

		// write operator
		operator := q.operatorToSQL(n)
		translator.SqlQuery.WriteString(fmt.Sprintf(" %s ", strings.ToUpper(operator)))

		q.TranslateToSQL(&n.Right, translator)
		translator.SqlQuery.WriteString(")")
	case *ast.NilNode:
		translator.SqlQuery.WriteString(fmt.Sprintf("%s", "NULL"))
	case *ast.IdentifierNode:
		translator.SqlQuery.WriteString(n.Value)
	case *ast.IntegerNode:
		translator.SqlQuery.WriteString(strconv.FormatInt(int64(n.Value), 10))
	case *ast.FloatNode:
		translator.SqlQuery.WriteString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	case *ast.BoolNode:
		translator.SqlQuery.WriteString(strconv.FormatBool(n.Value))
	case *ast.StringNode:
		translator.SqlQuery.WriteString(fmt.Sprintf("'%s'", n.Value))
	case *ast.ConstantNode:
		translator.SqlQuery.WriteString(fmt.Sprintf("%s", n.Value))
	case *ast.UnaryNode:
		switch n.Operator {
		case "not":
			binaryNode, ok := (n.Node).(*ast.BinaryNode)
			if ok && binaryNode.Operator == "in" {
				ast.Patch(&n.Node, &ast.BinaryNode{
					Operator: "not in",
					Left:     binaryNode.Left,
					Right:    binaryNode.Right,
				})
			}
		case "!":
			switch nodeV := n.Node.(type) {
			case *ast.BoolNode:
				ast.Patch(&n.Node, &ast.BoolNode{
					Value: !nodeV.Value,
				})
				// adjust other type if needed
			}
		}
		q.TranslateToSQL(&n.Node, translator)
	case *ast.BuiltinNode:
		result, err := q.getQueryExprResult(n.String())
		if err != nil {
			return
		}
		translator.SqlQuery.WriteString(fmt.Sprintf("%s", result))
	case *ast.ArrayNode:
		translator.SqlQuery.WriteString("(")
		for i := range n.Nodes {
			q.TranslateToSQL(&n.Nodes[i], translator)
			if i != len(n.Nodes)-1 {
				translator.SqlQuery.WriteString(", ")
			}
		}
		translator.SqlQuery.WriteString(")")
	case *ast.ConditionalNode:
		result, err := q.getQueryExprResult(n.String())
		if err != nil {
			return
		}
		if nodeV, ok := result.(ast.Node); ok {
			q.TranslateToSQL(&nodeV, translator)
		}
	case *ast.ChainNode:
	case *ast.MemberNode:
	case *ast.SliceNode:
	case *ast.CallNode:
	case *ast.ClosureNode:
	case *ast.PointerNode:
	case *ast.VariableDeclaratorNode:
	case *ast.MapNode:
	case *ast.PairNode:
	default:
		panic(fmt.Sprintf("undefined node type (%T)", node))
	}
}

func (q *QueryExprTranslator) getQueryExprResult(fn string) (any, error) {
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

func (q *QueryExprTranslator) operatorToSQL(bn *ast.BinaryNode) string {
	switch {
	case bn.Operator == "&&":
		return "AND"
	case bn.Operator == "||":
		return "OR"
	case bn.Operator == "!=":
		if _, ok := bn.Right.(*ast.NilNode); ok {
			return "IS NOT"
		}
	case bn.Operator == "==":
		if _, ok := bn.Right.(*ast.NilNode); ok {
			return "IS"
		} else {
			return "="
		}
	}

	return bn.Operator
}

func (q *QueryExprTranslator) ConvertToEsQuery() (string, error) {
	esQueryInterface := q.TranslateToEsQuery(q.getTreeNodeFromQueryExpr())
	esQuery, ok := esQueryInterface.(map[string]interface{})
	if !ok {
		return "", errors.New("failed to generate Elasticsearch query")
	}
	q.EsQuery = map[string]interface{}{"query": esQuery}

	// Convert to JSON
	queryJSON, err := json.Marshal(q.EsQuery)
	if err != nil {
		return "", err
	}

	return string(queryJSON), nil
}

// TranslateToEsQuery The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (q *QueryExprTranslator) TranslateToEsQuery(node *ast.Node) interface{} {
	if *node == nil {
		return nil
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		return q.translateBinaryNodeToEsQuery(n)
	case *ast.NilNode:
		return nil
	case *ast.IdentifierNode:
		return n.Value
	case *ast.IntegerNode:
		return n.Value
	case *ast.FloatNode:
		return n.Value
	case *ast.BoolNode:
		return n.Value
	case *ast.StringNode:
		return n.Value
	case *ast.UnaryNode:
		return q.translateUnaryNodeToEsQuery(n)
	case *ast.ArrayNode:
		return q.translateArrayNodeToEsQuery(n)
	case *ast.ConstantNode:
		return n.Value
	case *ast.BuiltinNode:
		result, err := q.getQueryExprResult(n.String())
		if err != nil {
			return nil
		}
		return result
	case *ast.ConditionalNode:
		result, err := q.getQueryExprResult(n.String())
		if err != nil {
			return nil
		}
		if nodeV, ok := result.(ast.Node); ok {
			return q.TranslateToEsQuery(&nodeV)
		}
	case *ast.ChainNode:
	case *ast.MemberNode:
	case *ast.SliceNode:
	case *ast.CallNode:
	case *ast.ClosureNode:
	case *ast.PointerNode:
	case *ast.VariableDeclaratorNode:
	case *ast.MapNode:
	case *ast.PairNode:
	default:
		panic(fmt.Sprintf("undefined node type (%T)", node))
	}

	return nil
}

func (q *QueryExprTranslator) translateBinaryNodeToEsQuery(n *ast.BinaryNode) map[string]interface{} {
	left := q.TranslateToEsQuery(&n.Left)
	right := q.TranslateToEsQuery(&n.Right)

	switch n.Operator {
	case "&&":
		return q.boolQuery("must", left, right)
	case "||":
		return q.boolQuery("should", left, right)
	case "==":
		return q.termQuery(left.(string), right)
	case "!=":
		return q.mustNotQuery(left.(string), right)
	case "<", "<=", ">", ">=":
		return q.rangeQuery(left.(string), q.operatorToEsQuery(n.Operator), right)
	case "in":
		return q.termsQuery(left.(string), right)
	default:
		return nil
	}
}

func (q *QueryExprTranslator) translateUnaryNodeToEsQuery(n *ast.UnaryNode) interface{} {
	switch n.Operator {
	case "not":
		if binaryNode, ok := n.Node.(*ast.BinaryNode); ok && binaryNode.Operator == "in" {
			left := q.TranslateToEsQuery(&binaryNode.Left)
			right := q.TranslateToEsQuery(&binaryNode.Right)
			return q.mustNotTermsQuery(left.(string), right)
		}
		return nil
	case "!":
		nodeValue := q.TranslateToEsQuery(&n.Node)
		switch value := nodeValue.(type) {
		case bool:
			return !value
		default:
			return map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []interface{}{nodeValue},
				},
			}
		}
	default:
		return nil
	}
}

func (q *QueryExprTranslator) translateArrayNodeToEsQuery(n *ast.ArrayNode) []interface{} {
	values := make([]interface{}, len(n.Nodes))
	for i, node := range n.Nodes {
		values[i] = q.TranslateToEsQuery(&node)
	}
	return values
}

func (q *QueryExprTranslator) operatorToEsQuery(operator string) string {
	switch operator {
	case ">":
		return "gt"
	case ">=":
		return "gte"
	case "<":
		return "lt"
	case "<=":
		return "lte"
	}

	return operator
}

func (q *QueryExprTranslator) boolQuery(condition string, left, right interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			condition: []interface{}{left, right},
		},
	}
}

func (q *QueryExprTranslator) termQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	}
}

func (q *QueryExprTranslator) mustNotQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must_not": []interface{}{
				map[string]interface{}{
					"term": map[string]interface{}{
						field: value,
					},
				},
			},
		},
	}
}

func (q *QueryExprTranslator) rangeQuery(field, operator string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"range": map[string]interface{}{
			field: map[string]interface{}{
				operator: value,
			},
		},
	}
}

func (q *QueryExprTranslator) termsQuery(field string, values interface{}) map[string]interface{} {
	return map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	}
}

func (q *QueryExprTranslator) mustNotTermsQuery(field string, values interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must_not": []interface{}{
				map[string]interface{}{
					"terms": map[string]interface{}{
						field: values,
					},
				},
			},
		},
	}
}
