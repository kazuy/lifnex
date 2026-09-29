package kawasaki

type eventsResponse struct {
	TotalCount int             `json:"total_numbers"`
	Events     []eventResponse `json:"event_data"`
}

type eventResponse struct {
	Title       string             `json:"title"`
	Content     string             `json:"content"`
	Dates       []dateResponse     `json:"date_list"`
	Place       string             `json:"place"`
	Address     string             `json:"place_adr"`
	Locations   []locationResponse `json:"event_location"`
	OpenURL     string             `json:"open_url"`
	RelatedURLs []relatedResponse  `json:"rel_list"`
}

type dateResponse struct {
	Date      string `json:"date"`
	StartTime string `json:"time_from"`
	EndTime   string `json:"time_to"`
	Details   string `json:"time_ext"`
}

type locationResponse struct {
	Area    string `json:"place"`
	Address string `json:"venue_address"`
}

type relatedResponse struct {
	URL string `json:"rel_url"`
}
