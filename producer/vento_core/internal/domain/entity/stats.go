package entity

type BusinessStats struct {
	TotalSales     float64 `json:"total_sales"`
	PendingOrders  int     `json:"pending_orders"`
	TotalProducts  int     `json:"total_products"`
	ConversionRate float64 `json:"conversion_rate"`
}
