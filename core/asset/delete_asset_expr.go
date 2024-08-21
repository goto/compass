package asset

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goto/compass/pkg/generichelper"
	"github.com/goto/compass/pkg/queryexpr"
)

var (
	assetJSONTagsSchema              = generichelper.GetJSONTags(Asset{})
	errTypeOrServiceHasWrongOperator = errors.New("identifier type and service must be equals (==) or IN operator")
	errMissRequiredIdentifier        = errors.New("must exists these identifiers: refreshed_at, type, and service")
)

type DeleteAssetExpr struct {
	queryexpr.ExprStr
}

func (d DeleteAssetExpr) ToQuery() (string, error) {
	return d.ExprStr.ToQuery()
}

func (d DeleteAssetExpr) Validate() error {
	identifiersWithOperator, err := queryexpr.GetIdentifiersMap(d.ExprStr.String())
	if err != nil {
		return err
	}

	if err := d.isRequiredIdentifiersExist(identifiersWithOperator); err != nil {
		return err
	}

	if err := d.isUsingRightOperator(identifiersWithOperator); err != nil {
		return err
	}

	return d.isAllIdentifiersExistInStruct(identifiersWithOperator)
}

func (DeleteAssetExpr) isRequiredIdentifiersExist(identifiersWithOperator map[string]string) error {
	isExist := func(jsonTag string) bool {
		return identifiersWithOperator[jsonTag] != ""
	}
	mustExist := isExist("refreshed_at") && isExist("type") && isExist("service")
	if !mustExist {
		return errMissRequiredIdentifier
	}
	return nil
}

func (DeleteAssetExpr) isUsingRightOperator(identifiersWithOperator map[string]string) error {
	isOperatorEqualsOrIn := func(jsonTag string) bool {
		return identifiersWithOperator[jsonTag] == "==" || strings.ToUpper(identifiersWithOperator[jsonTag]) == "IN"
	}
	if !isOperatorEqualsOrIn("type") || !isOperatorEqualsOrIn("service") {
		return errTypeOrServiceHasWrongOperator
	}
	return nil
}

func (DeleteAssetExpr) isAllIdentifiersExistInStruct(identifiersWithOperator map[string]string) error {
	identifiers := generichelper.GetMapKeys(identifiersWithOperator)
	for _, identifier := range identifiers {
		isFieldValid := generichelper.Contains(assetJSONTagsSchema, identifier)
		if !isFieldValid {
			return fmt.Errorf("%s is not a valid identifier", identifier)
		}
	}
	return nil
}
