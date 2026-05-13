package model

type Asset struct {
	Name      string `json:"name"`
	AssetType string `json:"assetType"`
	Resource  string `json:"resource,omitempty"`
}

type ExportAssetsRequest struct {
	AssetTypes    []string `json:"assetTypes,omitempty"`
	ContentType   string   `json:"contentType,omitempty"`
	OutputConfig  string   `json:"outputConfig,omitempty"`
}
