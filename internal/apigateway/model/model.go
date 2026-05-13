package model

type Gateway struct {
	Name     string `json:"name"`
	APIConfig string `json:"apiConfig"`
	DisplayName string `json:"displayName,omitempty"`
}

type APIConfig struct {
	Name      string `json:"name"`
	GatewayConfig string `json:"gatewayConfig,omitempty"`
}

type API struct {
	Name     string `json:"name"`
	APIConfig string `json:"apiConfig"`
}
