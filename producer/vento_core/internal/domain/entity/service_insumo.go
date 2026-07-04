package entity

import "time"

// ServiceInsumo declares that producing one unit of a Service consumes a
// fixed quantity of an Insumo (e.g. "Impresión 3D" consumes 1.2g de resina
// por unidad). Owned by the Service, mirrors how OrderPayment is owned by
// the Order.
type ServiceInsumo struct {
	ID              string    `json:"id"`
	ServiceID       string    `json:"service_id"`
	InsumoID        string    `json:"insumo_id"`
	QuantityPerUnit float64   `json:"quantity_per_unit"`
	CreatedAt       time.Time `json:"created_at"`
}

// ServiceInsumoDetail enriches a ServiceInsumo with display info from its
// Insumo (name, unit), used when listing a service's recipe for the catalog
// UI — same enrichment pattern as OrderPaymentDetail.
type ServiceInsumoDetail struct {
	ServiceInsumo
	InsumoName string `json:"insumo_name"`
	Unit       string `json:"unit"`
}
