package asset

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
)

var supportedTypes = []Type{
	typeTable,
	typeJob,
	typeDashboard,
	typeTopic,
	typeFeatureTable,
	typeApplication,
	typeModel,
	typeQuery,
	typeMetric,
}

var isTypeSupported = map[Type]bool{
	typeTable:        true,
	typeJob:          true,
	typeDashboard:    true,
	typeTopic:        true,
	typeFeatureTable: true,
	typeApplication:  true,
	typeModel:        true,
	typeQuery:        true,
	typeMetric:       true,
}

func GetSupportedTypes() []Type {
	output := make([]Type, 0, len(supportedTypes))
	output = append(output, supportedTypes...)
	return output
}

// Type specifies a supported type name
type Type string

// String cast Type to string
func (t Type) String() string {
	return string(t)
}

// IsValid will validate whether the typename is valid or not
func (t Type) IsValid() bool {
	return isTypeSupported[t]
}
