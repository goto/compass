package asset

import (
	"errors"
	"fmt"
)

const AllServicesCleanupConfig = "all"

var (
	ErrEmptyID                   = errors.New("asset does not have ID")
	ErrProbeExists               = errors.New("asset probe already exists")
	ErrEmptyURN                  = errors.New("asset does not have URN")
	ErrEmptyQuery                = errors.New("query is empty")
	ErrEmptyServices             = errors.New("services is empty")
	ErrUnknownType               = errors.New("unknown type")
	ErrNilAsset                  = errors.New("nil asset")
	ErrURNExist                  = errors.New("urn asset is already exist")
	ErrAssetAlreadyDeleted       = errors.New("asset already deleted")
	ErrExpiryThresholdTimeIsZero = errors.New("expiry threshold time is zero")
)

type NotFoundError struct {
	AssetID string
	URN     string
}

func (err NotFoundError) Error() string {
	if err.AssetID != "" {
		return fmt.Sprintf("no such record: %q", err.AssetID)
	} else if err.URN != "" {
		return fmt.Sprintf("could not find asset with urn = %s", err.URN)
	}

	return "could not find asset"
}

type InvalidError struct {
	AssetID string
}

func (err InvalidError) Error() string {
	return fmt.Sprintf("invalid asset id: %q", err.AssetID)
}

type DiscoveryError struct {
	Op     string
	ID     string
	Index  string
	ESCode string
	Err    error
}

func (err DiscoveryError) Error() string {
	msg := "discovery error: "
	if err.Op != "" {
		msg += err.Op + ": "
	}
	if err.ID != "" {
		msg += fmt.Sprintf("doc ID '%s': ", err.ID)
	}
	if err.Index != "" {
		msg += fmt.Sprintf("index '%s': ", err.Index)
	}
	if err.ESCode != "" {
		msg += fmt.Sprintf("elasticsearch code '%s': ", err.ESCode)
	}
	msg += err.Err.Error()
	return msg
}
