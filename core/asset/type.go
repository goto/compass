package asset

const (
	TypeTable        Type = "table"
	TypeJob          Type = "job"
	TypeDashboard    Type = "dashboard"
	TypeTopic        Type = "topic"
	TypeFeatureTable Type = "feature_table"
	TypeApplication  Type = "application"
	TypeModel        Type = "model"
	TypeQuery        Type = "query"
	TypeMetric       Type = "metric"
)

var supportedTypes = []Type{
	TypeTable,
	TypeJob,
	TypeDashboard,
	TypeTopic,
	TypeFeatureTable,
	TypeApplication,
	TypeModel,
	TypeQuery,
	TypeMetric,
}

var isTypeSupported = map[Type]bool{
	TypeTable:        true,
	TypeJob:          true,
	TypeDashboard:    true,
	TypeTopic:        true,
	TypeFeatureTable: true,
	TypeApplication:  true,
	TypeModel:        true,
	TypeQuery:        true,
	TypeMetric:       true,
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
