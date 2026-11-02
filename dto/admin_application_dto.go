package dto

type AdminApplicationListRequest struct {
	Status   string `query:"status"`
	Category string `query:"category"`
	Search   string `query:"search"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}

type AdminApplicationListItem struct {
	ID               uint   `json:"id"`
	RegistrationDate string `json:"registration_date"`
	ApplicantName    string `json:"applicant_name"`
	Category         string `json:"category"`
	Period           string `json:"period"`
	Status           string `json:"status"`
}

type AdminApplicationListResponse struct {
	Total int                        `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
	Items []AdminApplicationListItem `json:"items"`
}