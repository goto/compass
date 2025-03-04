package queryexpr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr/ast"
)

type SQLExpr string

func (s SQLExpr) String() string {
	return string(s)
}

// ToQuery default: convert to be SQL query
func (s SQLExpr) ToQuery() (string, error) {
	queryExprParsed, err := getTreeNodeFromQueryExpr(s.String())
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
func (SQLExpr) Validate() error {
	return nil
}

// convertToSQL The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (s SQLExpr) convertToSQL(node ast.Node, stringBuilder *strings.Builder) error {
	if node == nil {
		return errCannotConvertNilQuery
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
		if err := s.getQueryExprResultForSQL(n.String(), stringBuilder); err != nil {
			return err
		}
	case *ast.MemberNode:
		memberIdentifiers := strings.Split(n.String(), ".")
		identifier := memberIdentifiers[0]
		for i := 1; i <= len(memberIdentifiers)-2; i++ {
			identifier += fmt.Sprintf("->'%s'", memberIdentifiers[i])
		}
		identifier += fmt.Sprintf("->>'%s'", memberIdentifiers[len(memberIdentifiers)-1])
		stringBuilder.WriteString(identifier)
	default:
		return s.unsupportedQueryError(n)
	}

	return nil
}

func (s SQLExpr) binaryNodeToSQLQuery(n *ast.BinaryNode, stringBuilder *strings.Builder) error {
	operator := s.operatorToSQL(n)
	if operator == "" { // most likely the node is an operation
		if err := s.getQueryExprResultForSQL(n.String(), stringBuilder); err != nil {
			return err
		}
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

// getQueryExprResultForSQL using getQueryExprResult to get the result of query expr operation, and make it as SQL syntax
func (SQLExpr) getQueryExprResultForSQL(queryExprOperation string, stringBuilder *strings.Builder) error {
	result, err := getQueryExprResult(queryExprOperation)
	if err != nil {
		return err
	}
	if str, ok := result.(string); ok {
		result = fmt.Sprintf("'%s'", str)
	}

	fmt.Fprintf(stringBuilder, "%v", result)
	return nil
}

func (s SQLExpr) arrayNodeToSQLQuery(n *ast.ArrayNode, stringBuilder *strings.Builder) error {
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

func (s SQLExpr) patchUnaryNode(n *ast.UnaryNode) error {
	switch n.Operator {
	case "not":
		binaryNode, ok := (n.Node).(*ast.BinaryNode)
		if !ok {
			return s.unsupportedQueryError(n)
		}
		if strings.ToUpper(binaryNode.Operator) == "IN" {
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
		default:
			result, err := getQueryExprResult(n.String())
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

func (SQLExpr) operatorToSQL(bn *ast.BinaryNode) string {
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

func (SQLExpr) unsupportedQueryError(node ast.Node) error {
	return fmt.Errorf("unsupported query expr: %s to SQL query", node.String())
}
