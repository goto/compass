package asset

import (
	"errors"
	"regexp"
)

var (
	errTypeInvalidLength    = errors.New("type length must be 3 to 16 inclusive")
	errTypeInvalidCharacter = errors.New("type must be combination of alphanumeric and underscores")
)

var invalidTypePattern = regexp.MustCompile(`[^a-z0-9_]`)

const (
	typeMinLength = 3
	typeMaxLength = 16
)

const (
	typeTable        Type = "table"
	typeJob          Type = "job"
	typeDashboard    Type = "dashboard"
	typeTopic        Type = "topic"
	typeFeatureTable Type = "feature_table"
	typeApplication  Type = "application"
	typeModel        Type = "model"
	typeQuery        Type = "query"
	typeMetric       Type = "metric"
	typeExperiment   Type = "experiment"
)

var supportedTypeMap = map[Type]bool{
	typeTable:        true,
	typeJob:          true,
	typeDashboard:    true,
	typeTopic:        true,
	typeFeatureTable: true,
	typeApplication:  true,
	typeModel:        true,
	typeQuery:        true,
	typeMetric:       true,
	typeExperiment:   true,
}

func GetSupportedTypes() []Type {
	output := make([]Type, 0, len(supportedTypeMap))
	for _type := range supportedTypeMap {
		output = append(output, _type)
	}
	return output
}

func RegisterSupportedTypes(types ...Type) error {
	for _, t := range types {
		if err := t.validate(); err != nil {
			return err
		}
	}

	for _, t := range types {
		if supported := supportedTypeMap[t]; !supported {
			supportedTypeMap[t] = true
		}
	}

	return nil
}

// Type specifies a supported type name
type Type string

// String cast Type to string
func (t Type) String() string {
	return string(t)
}

// IsValid will validate whether the typename is valid or not
func (t Type) IsValid() bool {
	return supportedTypeMap[t]
}

func (t Type) validate() error {
	if l := len(t.String()); l < typeMinLength || l > typeMaxLength {
		return errTypeInvalidLength
	}

	if invalidTypePattern.FindStringSubmatch(t.String()) != nil {
		return errTypeInvalidCharacter
	}

	return nil
}
