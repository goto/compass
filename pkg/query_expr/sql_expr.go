package queryexpr

import (
	"fmt"
	"github.com/expr-lang/expr/ast"
	"strconv"
	"strings"
)

type SQLExpr struct {
	QueryExpr string
	SQLQuery  strings.Builder
}

// ToQuery default
func (s *SQLExpr) ToQuery() (string, error) {
	queryExprParsed, err := GetTreeNodeFromQueryExpr(s.QueryExpr)
	if err != nil {
		return "", err
	}
	s.ConvertToSQL(queryExprParsed)
	return s.SQLQuery.String(), nil
}

// Validate default: no validation
func (*SQLExpr) Validate() error {
	return nil
}

// ConvertToSQL The idea came from ast.Walk. Currently, the development focus implement for the node type that most likely used in our needs.
// TODO: implement translator for node type that still not covered right now.
func (s *SQLExpr) ConvertToSQL(node *ast.Node) {
	if *node == nil {
		return
	}
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		s.SQLQuery.WriteString("(")
		s.ConvertToSQL(&n.Left)

		// write operator
		operator := s.operatorToSQL(n)
		s.SQLQuery.WriteString(fmt.Sprintf(" %s ", strings.ToUpper(operator)))

		s.ConvertToSQL(&n.Right)
		s.SQLQuery.WriteString(")")
	case *ast.NilNode:
		s.SQLQuery.WriteString("NULL")
	case *ast.IdentifierNode:
		s.SQLQuery.WriteString(n.Value)
	case *ast.IntegerNode:
		s.SQLQuery.WriteString(strconv.FormatInt(int64(n.Value), 10))
	case *ast.FloatNode:
		s.SQLQuery.WriteString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	case *ast.BoolNode:
		s.SQLQuery.WriteString(strconv.FormatBool(n.Value))
	case *ast.StringNode:
		s.SQLQuery.WriteString(fmt.Sprintf("'%s'", n.Value))
	case *ast.ConstantNode:
		s.SQLQuery.WriteString(fmt.Sprintf("%s", n.Value))
	case *ast.UnaryNode:
		s.patchUnaryNode(n)
		s.ConvertToSQL(&n.Node)
	case *ast.BuiltinNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return
		}
		s.SQLQuery.WriteString(fmt.Sprintf("%s", result))
	case *ast.ArrayNode:
		s.SQLQuery.WriteString("(")
		for i := range n.Nodes {
			s.ConvertToSQL(&n.Nodes[i])
			if i != len(n.Nodes)-1 {
				s.SQLQuery.WriteString(", ")
			}
		}
		s.SQLQuery.WriteString(")")
	case *ast.ConditionalNode:
		result, err := GetQueryExprResult(n.String())
		if err != nil {
			return
		}
		if nodeV, ok := result.(ast.Node); ok {
			s.ConvertToSQL(&nodeV)
		}
	}
}

func (*SQLExpr) patchUnaryNode(n *ast.UnaryNode) {
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
			// TODO: adjust other types if needed
		}
	}
}

func (*SQLExpr) operatorToSQL(bn *ast.BinaryNode) string {
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
		}
		return "="
	}

	return bn.Operator
}
