package dto

// MetricsResponse is the full payload returned by GET /metrics.
type MetricsResponse struct {
	SalesByDay             []SalesByDayPoint `json:"sales_by_day"`
	TopProducts            []TopProduct      `json:"top_products"`
	ConversationsByChannel []ChannelCount    `json:"conversations_by_channel"`
	ConversionFunnel       []FunnelStep      `json:"conversion_funnel"`
	Summary                MetricsSummary    `json:"summary"`
}

type SalesByDayPoint struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type TopProduct struct {
	Name    string  `json:"name"`
	Units   int     `json:"units"`
	Revenue float64 `json:"revenue"`
}

type ChannelCount struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type FunnelStep struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type MetricsSummary struct {
	TotalRevenue        float64 `json:"total_revenue"`
	RevenueChange       float64 `json:"revenue_change"`
	TotalOrders         int     `json:"total_orders"`
	OrdersChange        float64 `json:"orders_change"`
	TotalConversations  int     `json:"total_conversations"`
	ConversationsChange float64 `json:"conversations_change"`
	AvgTicket           float64 `json:"avg_ticket"`
	TicketChange        float64 `json:"ticket_change"`
}
