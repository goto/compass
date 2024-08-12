package queryexpr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr/ast"
)

type SQLExpr string

func (s *SQLExpr) String() string {
	return string(*s)
}

// ToQuery default
func (s *SQLExpr) ToQuery() (string, error) {
	queryExprParsed, err := GetTreeNodeFromQueryExpr(s.String())
	if err != nil {
		return "", err
	}
	stringBuilder := &strings.Builder{}
	if err := s.convertToSQL(queryExprParsed, stringBuilder); err != nil {
		return "", err
	}
	return stringBuilder.String(), nil
}

// Validate default: no validation
func (*SQLExpr) Validate() error {
	return nil
}

// convertToSQL The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (s *SQLExpr) convertToSQL(node ast.Node, stringBuilder *strings.Builder) error {
	if node == nil {
		return fmt.Errorf("cannot convert nil to SQL query")
	}
	switch n := (node).(type) {
	case *ast.BinaryNode:
		err := s.binaryNodeToSQLQuery(n, stringBuilder)
		if err != nil {
			return err
		}
	case *ast.NilNode:
		stringBuilder.WriteString("NULL")
	case *ast.IdentifierNode:
		stringBuilder.WriteString(n.Value)
	case *ast.IntegerNode:
		stringBuilder.WriteString(strconv.FormatInt(int64(n.Value), 10))
	case *ast.FloatNode:
		stringBuilder.WriteString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	case *ast.BoolNode:
		stringBuilder.WriteString(strconv.FormatBool(n.Value))
	case *ast.StringNode:
		fmt.Fprintf(stringBuilder, "'%s'", n.Value)
	case *ast.ConstantNode:
		fmt.Fprintf(stringBuilder, "%v", n.Value)
	case *ast.UnaryNode:
		if err := s.patchUnaryNode(n); err != nil {
			return err
		}
		if err := s.convertToSQL(n.Node, stringBuilder); err != nil {
			return err
		}
	case *ast.ArrayNode:
		err := s.arrayNodeToSQLQuery(n, stringBuilder)
		if err != nil {
			return err
		}
	case *ast.BuiltinNode, *ast.ConditionalNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return err
		}
		fmt.Fprintf(stringBuilder, "%v", result)
	default:
		return s.unsupportedQueryError(n)
	}

	return nil
}

func (s *SQLExpr) binaryNodeToSQLQuery(n *ast.BinaryNode, stringBuilder *strings.Builder) error {
	operator := s.operatorToSQL(n)
	if operator == "" { // most likely the node is operation
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return err
		}
		fmt.Fprintf(stringBuilder, "%v", result)
	} else {
		stringBuilder.WriteString("(")
		if err := s.convertToSQL(n.Left, stringBuilder); err != nil {
			return err
		}

		// write operator
		fmt.Fprintf(stringBuilder, " %s ", strings.ToUpper(operator))

		if err := s.convertToSQL(n.Right, stringBuilder); err != nil {
			return err
		}
		stringBuilder.WriteString(")")
	}

	return nil
}

func (s *SQLExpr) arrayNodeToSQLQuery(n *ast.ArrayNode, stringBuilder *strings.Builder) error {
	stringBuilder.WriteString("(")
	for i := range n.Nodes {
		if err := s.convertToSQL(n.Nodes[i], stringBuilder); err != nil {
			return err
		}
		if i != len(n.Nodes)-1 {
			stringBuilder.WriteString(", ")
		}
	}
	stringBuilder.WriteString(")")
	return nil
}

func (s *SQLExpr) patchUnaryNode(n *ast.UnaryNode) error {
	switch n.Operator {
	case "not":
		binaryNode, ok := (n.Node).(*ast.BinaryNode)
		if ok && strings.ToUpper(binaryNode.Operator) == "IN" {
			ast.Patch(&n.Node, &ast.BinaryNode{
				Operator: "not in",
				Left:     binaryNode.Left,
				Right:    binaryNode.Right,
			})
		} else {
			return s.unsupportedQueryError(n)
		}
	case "!":
		switch nodeV := n.Node.(type) {
		case *ast.BoolNode:
			ast.Patch(&n.Node, &ast.BoolNode{
				Value: !nodeV.Value,
			})
		default:
			result, err := GetQueryExprResult(n.String())
			if err != nil {
				return err
			}
			if boolResult, ok := result.(bool); ok {
				ast.Patch(&n.Node, &ast.BoolNode{
					Value: boolResult,
				})
				return nil
			}
			return s.unsupportedQueryError(n)
		}
	}

	return nil
}

func (*SQLExpr) operatorToSQL(bn *ast.BinaryNode) string {
	switch strings.ToUpper(bn.Operator) {
	case "&&":
		return "AND"
	case "||":
		return "OR"
	case "!=":
		if _, ok := bn.Right.(*ast.NilNode); ok {
			return "IS NOT"
		}
		return bn.Operator
	case "==":
		if _, ok := bn.Right.(*ast.NilNode); ok {
			return "IS"
		}
		return "="
	case "<", "<=", ">", ">=":
		return bn.Operator
	case "IN", "NOT IN":
		return bn.Operator
	}

	return "" // identify operation, like: +, -, *, etc
}

func (*SQLExpr) unsupportedQueryError(node ast.Node) error {
	return fmt.Errorf("unsupported query expr: %s to SQL query", node.String())
}
