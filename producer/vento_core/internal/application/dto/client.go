package dto

type CreateClientRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	SocialNetwork string `json:"social_network"`
	SocialHandle  string `json:"social_handle"`
}

type ClientResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Phone             string  `json:"phone"`
	SocialNetwork     string  `json:"social_network"`
	SocialHandle      string  `json:"social_handle"`
	LastPurchaseDate  string  `json:"last_purchase_date"`
	DaysSincePurchase int     `json:"days_since_purchase"`
	TotalSpent        float64 `json:"total_spent"`
	OrderCount        int     `json:"order_count"`
	Status            string  `json:"status"`
}
