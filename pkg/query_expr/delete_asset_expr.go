package queryexpr

import (
	"fmt"
	"strings"

	"github.com/goto/compass/core/asset"
	generichelper "github.com/goto/compass/pkg/generic_helper"
)

type DeleteAssetExpr struct {
	ExprStr
}

func (d DeleteAssetExpr) ToQuery() (string, error) {
	return d.ExprStr.ToQuery()
}

func (d DeleteAssetExpr) Validate() error {
	identifiersWithOperator, err := GetIdentifiersMap(d.ExprStr.String())
	if err != nil {
		return err
	}

	isExist := func(jsonTag string) bool {
		return identifiersWithOperator[jsonTag] != ""
	}
	mustExist := isExist("refreshed_at") && isExist("type") && isExist("service")
	if !mustExist {
		return fmt.Errorf("must exists these identifiers: refreshed_at, type, and service")
	}

	isOperatorEqualsOrIn := func(jsonTag string) bool {
		return identifiersWithOperator[jsonTag] == "==" || strings.ToUpper(identifiersWithOperator[jsonTag]) == "IN"
	}
	if !isOperatorEqualsOrIn("type") || !isOperatorEqualsOrIn("service") {
		return fmt.Errorf("identifier type and service must be equals (==) or IN operator")
	}

	identifiers := generichelper.GetMapKeys(identifiersWithOperator)
	jsonTagsSchema := generichelper.GetJSONTags(asset.Asset{})
	for _, identifier := range identifiers {
		isFieldValid := generichelper.Contains(jsonTagsSchema, identifier)
		if !isFieldValid {
			return fmt.Errorf("%s is not a valid identifier", identifier)
		}
	}

	return nil
}
