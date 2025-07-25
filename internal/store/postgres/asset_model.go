package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/jmoiron/sqlx/types"
	"github.com/r3labs/diff/v2"
)

type AssetModel struct {
	ID          string     `db:"id"`
	URN         string     `db:"urn"`
	Type        string     `db:"type"`
	Name        string     `db:"name"`
	Service     string     `db:"service"`
	Description string     `db:"description"`
	Data        JSONMap    `db:"data"`
	URL         string     `db:"url"`
	Labels      JSONMap    `db:"labels"`
	IsDeleted   bool       `db:"is_deleted"`
	Version     string     `db:"version"`
	UpdatedBy   UserModel  `db:"updated_by"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	RefreshedAt *time.Time `db:"refreshed_at"`
	// version specific information
	Changelog types.JSONText `db:"changelog"`
	Owners    types.JSONText `db:"owners"`
}

func (a *AssetModel) toAsset(owners []user.User) asset.Asset {
	return asset.Asset{
		ID:          a.ID,
		URN:         a.URN,
		Type:        asset.Type(a.Type),
		Name:        a.Name,
		Service:     a.Service,
		Description: a.Description,
		Data:        a.Data,
		URL:         a.URL,
		Labels:      a.buildLabels(),
		IsDeleted:   a.IsDeleted,
		Owners:      owners,
		Version:     a.Version,
		UpdatedBy:   a.UpdatedBy.toUser(),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		RefreshedAt: a.RefreshedAt,
	}
}

func (a *AssetModel) toAssetVersion() (asset.Asset, error) {
	var clog diff.Changelog
	err := a.Changelog.Unmarshal(&clog)
	if err != nil {
		return asset.Asset{}, err
	}

	return asset.Asset{
		ID:          a.ID,
		URN:         a.URN,
		Type:        asset.Type(a.Type),
		Service:     a.Service,
		IsDeleted:   a.IsDeleted,
		Version:     a.Version,
		Changelog:   clog,
		UpdatedBy:   a.UpdatedBy.toUser(),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		RefreshedAt: a.RefreshedAt,
	}, nil
}

func (a *AssetModel) toVersionedAsset(latestAssetVersion asset.Asset) (asset.Asset, error) {
	var owners []user.User
	err := a.Owners.Unmarshal(&owners)
	if err != nil {
		return asset.Asset{}, err
	}

	var clog diff.Changelog
	err = a.Changelog.Unmarshal(&clog)
	if err != nil {
		return asset.Asset{}, err
	}

	return asset.Asset{
		ID:          latestAssetVersion.ID,
		URN:         latestAssetVersion.URN,
		Type:        latestAssetVersion.Type,
		Name:        a.Name,
		Service:     latestAssetVersion.Service,
		Description: a.Description,
		Data:        a.Data,
		Labels:      a.buildLabels(),
		IsDeleted:   a.IsDeleted,
		Owners:      owners,
		Version:     a.Version,
		UpdatedBy:   a.UpdatedBy.toUser(),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		RefreshedAt: a.RefreshedAt,
		Changelog:   clog,
	}, nil
}

func (a *AssetModel) buildLabels() map[string]string {
	if a.Labels == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range a.Labels {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		result[strKey] = strValue
	}

	return result
}

type AssetProbeModel struct {
	ID           string    `db:"id"`
	AssetURN     string    `db:"asset_urn"`
	Status       string    `db:"status"`
	StatusReason string    `db:"status_reason"`
	Metadata     JSONMap   `db:"metadata"`
	Timestamp    time.Time `db:"timestamp"`
	CreatedAt    time.Time `db:"created_at"`
}

func (m *AssetProbeModel) toAssetProbe() asset.Probe {
	return asset.Probe{
		ID:           m.ID,
		AssetURN:     m.AssetURN,
		Status:       m.Status,
		StatusReason: m.StatusReason,
		Metadata:     m.Metadata,
		Timestamp:    m.Timestamp,
		CreatedAt:    m.CreatedAt,
	}
}

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	ba, err := m.MarshalJSON()
	return string(ba), err
}

func (m *JSONMap) Scan(value interface{}) error {
	var ba []byte
	switch v := value.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	t := map[string]interface{}{}
	err := json.Unmarshal(ba, &t)
	*m = JSONMap(t)
	return err
}

// MarshalJSON to output non base64 encoded []byte
func (m JSONMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	t := (map[string]interface{})(m)
	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (m *JSONMap) UnmarshalJSON(b []byte) error {
	t := map[string]interface{}{}
	err := json.Unmarshal(b, &t)
	*m = JSONMap(t)
	return err
}
