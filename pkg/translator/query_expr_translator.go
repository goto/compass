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

type ExprParam map[string]interface{}

type QueryExprTranslator struct {
	QueryExpr   string
	SqlQuery    strings.Builder
	EsQuery     map[string]interface{}
	Identifiers []string
}

func (q *QueryExprTranslator) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		q.Identifiers = append(q.Identifiers, n.Value)
	}
}

func (q *QueryExprTranslator) GetIdentifiers() []string {
	ast.Walk(q.getTreeNodeFromQueryExpr(), q)
	return q.Identifiers
}

func (q *QueryExprTranslator) getTreeNodeFromQueryExpr() *ast.Node {
	parsed, err := parser.Parse(q.QueryExpr)
	if err != nil {
		log.Fatalf("Error parsing expression: %v", err)
	}

	return &parsed.Node
}

func (q *QueryExprTranslator) ConvertToSQL() (string, error) {
	q.SqlQuery = strings.Builder{}
	q.translateToSQL(q.getTreeNodeFromQueryExpr(), q)
	return q.SqlQuery.String(), nil
}

// translateToSQL The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (q *QueryExprTranslator) translateToSQL(node *ast.Node, translator *QueryExprTranslator) {
	if *node == nil {
		return
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		translator.SqlQuery.WriteString("(")
		q.translateToSQL(&n.Left, translator)

		// write operator
		operator := q.operatorToSQL(n)
		translator.SqlQuery.WriteString(fmt.Sprintf(" %s ", strings.ToUpper(operator)))

		q.translateToSQL(&n.Right, translator)
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
		q.translateToSQL(&n.Node, translator)
	case *ast.BuiltinNode:
		result, err := q.getQueryExprResult(n.String())
		if err != nil {
			return
		}
		translator.SqlQuery.WriteString(fmt.Sprintf("%s", result))
	case *ast.ArrayNode:
		translator.SqlQuery.WriteString("(")
		for i := range n.Nodes {
			q.translateToSQL(&n.Nodes[i], translator)
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
			q.translateToSQL(&nodeV, translator)
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
	esQueryInterface := q.translateToEsQuery(q.getTreeNodeFromQueryExpr())
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

// translateToEsQuery The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (q *QueryExprTranslator) translateToEsQuery(node *ast.Node) interface{} {
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
			return q.translateToEsQuery(&nodeV)
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
	left := q.translateToEsQuery(&n.Left)
	right := q.translateToEsQuery(&n.Right)

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
			left := q.translateToEsQuery(&binaryNode.Left)
			right := q.translateToEsQuery(&binaryNode.Right)
			return q.mustNotTermsQuery(left.(string), right)
		}
		return nil
	case "!":
		nodeValue := q.translateToEsQuery(&n.Node)
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
		values[i] = q.translateToEsQuery(&node)
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
