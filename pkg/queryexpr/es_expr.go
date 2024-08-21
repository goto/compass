package queryexpr

import (
	"encoding/json"
	"fmt"

	"github.com/expr-lang/expr/ast"
)

var KeywordIdentifiers = [...]string{"service"}

type ESExpr string

func (e ESExpr) String() string {
	return string(e)
}

// ToQuery default: convert to be Elasticsearch query
func (e ESExpr) ToQuery() (string, error) {
	queryExprParsed, err := getTreeNodeFromQueryExpr(e.String())
	if err != nil {
		return "", err
	}

	esQueryInterface, err := e.translateToEsQuery(queryExprParsed)
	if err != nil {
		return "", err
	}
	esQuery, ok := esQueryInterface.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to generate Elasticsearch query")
	}
	esQuery = map[string]interface{}{"query": esQuery}

	queryJSON, err := json.Marshal(esQuery)
	if err != nil {
		return "", err
	}

	return string(queryJSON), nil
}

// Validate default: no validation
func (ESExpr) Validate() error {
	return nil
}

// translateToEsQuery The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (e ESExpr) translateToEsQuery(node ast.Node) (interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot convert nil to Elasticsearch query")
	}
	switch n := (node).(type) {
	case *ast.BinaryNode:
		return e.binaryNodeToEsQuery(n)
	case *ast.NilNode:
		return nil, nil
	case *ast.IdentifierNode:
		if e.isKeywordIdentifier(n) {
			return fmt.Sprintf("%s.keyword", n.Value), nil
		}
		return n.Value, nil
	case *ast.IntegerNode:
		return n.Value, nil
	case *ast.FloatNode:
		return n.Value, nil
	case *ast.BoolNode:
		return n.Value, nil
	case *ast.StringNode:
		return n.Value, nil
	case *ast.UnaryNode:
		return e.unaryNodeToEsQuery(n)
	case *ast.ArrayNode:
		return e.arrayNodeToEsQuery(n)
	case *ast.ConstantNode:
		return n.Value, nil
	case *ast.BuiltinNode, *ast.ConditionalNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, e.unsupportedQueryError(n)
	}
}

func (e ESExpr) binaryNodeToEsQuery(n *ast.BinaryNode) (interface{}, error) { //nolint:gocognit
	left, err := e.translateToEsQuery(n.Left)
	if err != nil {
		return nil, err
	}
	right, err := e.translateToEsQuery(n.Right)
	if err != nil {
		return nil, err
	}

	switch n.Operator {
	case "&&":
		return e.boolQuery("must", left, right), nil

	case "||":
		return e.boolQuery("should", left, right), nil

	case "==":
		if leftStr, ok := left.(string); ok {
			return e.termQuery(leftStr, right), nil
		}
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil

	case "!=":
		if leftStr, ok := left.(string); ok {
			return e.mustNotQuery(leftStr, right), nil
		}
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil

	case "<", "<=", ">", ">=":
		if leftStr, ok := left.(string); ok {
			return e.rangeQuery(leftStr, e.operatorToEsQuery(n.Operator), right), nil
		}
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil

	case "in":
		if leftStr, ok := left.(string); ok {
			return e.termsQuery(leftStr, right), nil
		}
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil

	default:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

func (ESExpr) isKeywordIdentifier(node *ast.IdentifierNode) bool {
	for _, keyword := range KeywordIdentifiers {
		if node.Value == keyword {
			return true
		}
	}
	return false
}

func (e ESExpr) unaryNodeToEsQuery(n *ast.UnaryNode) (interface{}, error) {
	switch n.Operator {
	case "not":
		if binaryNode, ok := n.Node.(*ast.BinaryNode); ok && binaryNode.Operator == "in" {
			left, err := e.translateToEsQuery(binaryNode.Left)
			if err != nil {
				return nil, err
			}
			right, err := e.translateToEsQuery(binaryNode.Right)
			if err != nil {
				return nil, err
			}
			return e.mustNotTermsQuery(left.(string), right), nil
		}
		return nil, e.unsupportedQueryError(n)

	case "!":
		nodeValue, err := e.translateToEsQuery(n.Node)
		if err != nil {
			return nil, err
		}
		switch value := nodeValue.(type) {
		case bool:
			return !value, nil
		default:
			return map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []interface{}{nodeValue},
				},
			}, nil
		}

	default:
		return nil, e.unsupportedQueryError(n)
	}
}

func (e ESExpr) arrayNodeToEsQuery(n *ast.ArrayNode) ([]interface{}, error) {
	values := make([]interface{}, len(n.Nodes))
	for i, node := range n.Nodes {
		nodeValue, err := e.translateToEsQuery(node)
		if err != nil {
			return nil, err
		}
		values[i] = nodeValue
	}
	return values, nil
}

func (ESExpr) operatorToEsQuery(operator string) string {
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

func (ESExpr) boolQuery(condition string, left, right interface{}) map[string]interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			condition: []interface{}{left, right},
		},
	}
}

func (ESExpr) termQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	}
}

func (ESExpr) mustNotQuery(field string, value interface{}) map[string]interface{} {
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

func (ESExpr) rangeQuery(field, operator string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"range": map[string]interface{}{
			field: map[string]interface{}{
				operator: value,
			},
		},
	}
}

func (ESExpr) termsQuery(field string, values interface{}) map[string]interface{} {
	return map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	}
}

func (ESExpr) mustNotTermsQuery(field string, values interface{}) map[string]interface{} {
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

func (ESExpr) unsupportedQueryError(node ast.Node) error {
	return fmt.Errorf("unsupported query expr: %s to Elasticsearch query", node.String())
}
