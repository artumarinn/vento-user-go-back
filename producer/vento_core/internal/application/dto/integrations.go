package dto

// IntegrationsStatusResponse reports which channels have a real, persisted
// configuration for the authenticated user. MercadoPago is always false
// because payment link integration is not built yet — never fake it.
type IntegrationsStatusResponse struct {
	WhatsApp    bool `json:"whatsapp"`
	Instagram   bool `json:"instagram"`
	Facebook    bool `json:"facebook"`
	MercadoPago bool `json:"mercadopago"`
	Connected   int  `json:"connected"`
	Total       int  `json:"total"`
}
