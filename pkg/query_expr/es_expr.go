package queryexpr

import (
	"encoding/json"
	"fmt"
	"github.com/expr-lang/expr/ast"
)

type ESExpr struct {
	QueryExpr string
	ESQuery   map[string]interface{}
}

// ToQuery default
func (e *ESExpr) ToQuery() (string, error) {
	queryExprParsed, err := GetTreeNodeFromQueryExpr(e.QueryExpr)
	if err != nil {
		return "", err
	}

	esQueryInterface := e.translateToEsQuery(queryExprParsed)
	esQuery, ok := esQueryInterface.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to generate Elasticsearch query")
	}
	e.ESQuery = map[string]interface{}{"query": esQuery}

	// Convert to JSON
	queryJSON, err := json.Marshal(e.ESQuery)
	if err != nil {
		return "", err
	}

	return string(queryJSON), nil
}

// Validate default: no validation
func (*ESExpr) Validate() error {
	return nil
}

// translateToEsQuery The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (e *ESExpr) translateToEsQuery(node *ast.Node) interface{} {
	if *node == nil {
		return nil
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		return e.translateBinaryNodeToEsQuery(n)
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
		return e.translateUnaryNodeToEsQuery(n)
	case *ast.ArrayNode:
		return e.translateArrayNodeToEsQuery(n)
	case *ast.ConstantNode:
		return n.Value
	case *ast.BuiltinNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil
		}
		return result
	case *ast.ConditionalNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil
		}
		if nodeV, ok := result.(ast.Node); ok {
			return e.translateToEsQuery(&nodeV)
		}
	}

	return nil
}

func (e *ESExpr) translateBinaryNodeToEsQuery(n *ast.BinaryNode) map[string]interface{} {
	left := e.translateToEsQuery(&n.Left)
	right := e.translateToEsQuery(&n.Right)

	switch n.Operator {
	case "&&":
		return e.boolQuery("must", left, right)
	case "||":
		return e.boolQuery("should", left, right)
	case "==":
		return e.termQuery(left.(string), right)
	case "!=":
		return e.mustNotQuery(left.(string), right)
	case "<", "<=", ">", ">=":
		return e.rangeQuery(left.(string), e.operatorToEsQuery(n.Operator), right)
	case "in":
		return e.termsQuery(left.(string), right)
	default:
		return nil
	}
}

func (e *ESExpr) translateUnaryNodeToEsQuery(n *ast.UnaryNode) interface{} {
	switch n.Operator {
	case "not":
		if binaryNode, ok := n.Node.(*ast.BinaryNode); ok && binaryNode.Operator == "in" {
			left := e.translateToEsQuery(&binaryNode.Left)
			right := e.translateToEsQuery(&binaryNode.Right)
			return e.mustNotTermsQuery(left.(string), right)
		}
		return nil
	case "!":
		nodeValue := e.translateToEsQuery(&n.Node)
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

func (e *ESExpr) translateArrayNodeToEsQuery(n *ast.ArrayNode) []interface{} {
	values := make([]interface{}, len(n.Nodes))
	for i, node := range n.Nodes {
		values[i] = e.translateToEsQuery(&node)
	}
	return values
}

func (*ESExpr) operatorToEsQuery(operator string) string {
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

func (*ESExpr) boolQuery(condition string, left, right interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			condition: []interface{}{left, right},
		},
	}
}

func (*ESExpr) termQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	}
}

func (*ESExpr) mustNotQuery(field string, value interface{}) map[string]interface{} {
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

func (*ESExpr) rangeQuery(field, operator string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"range": map[string]interface{}{
			field: map[string]interface{}{
				operator: value,
			},
		},
	}
}

func (*ESExpr) termsQuery(field string, values interface{}) map[string]interface{} {
	return map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	}
}

func (*ESExpr) mustNotTermsQuery(field string, values interface{}) map[string]interface{} {
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
