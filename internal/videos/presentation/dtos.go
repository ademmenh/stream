package presentation

type CreateVideoDto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type TriggerProcessingDto struct {
	RequestedQualities []string `json:"requested_qualities"`
}

type RegenerateQualityDto struct {
	Quality string `json:"quality"`
}

type ListVideosQuery struct {
	Search string  `query:"search"`
	Type   *string `query:"type"`
	Status *string `query:"status"`
	Page   int     `query:"page"`
	Limit  int     `query:"limit"`
	SortBy string  `query:"sort_by"`
	Order  string  `query:"order"`
}

type ListCatalogQuery struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}
