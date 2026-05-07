package asset

import "context"

//go:generate mockery --name=LineageParserClient -r --case underscore --with-expecter --structname LineageParserClient --filename lineage_parser_client_mock.go --output=./mocks

// LineageParserClient fetches lineage from an external service.
type LineageParserClient interface {
	FetchColumnLineage(ctx context.Context, query string) (LineageGraph, error)
}
