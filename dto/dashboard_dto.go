package dto

type DashboardStats struct {
	PendaftarSMK int64 `json:"pendaftar_smk"`
	PendaftarD3  int64 `json:"pendaftar_d3"`
	PendaftarD4  int64 `json:"pendaftar_d4"`
	PendaftarS1  int64 `json:"pendaftar_s1"`
}

type DashboardChartYear struct {
	Year     int   `json:"year"`
	Total    int64 `json:"total"`
	Diterima int64 `json:"diterima"`
	Ditolak  int64 `json:"ditolak"`
}

type RecentApplicationItem struct {
	ID               uint   `json:"id"`
	RegistrationDate string `json:"registration_date"`
	ApplicantName    string `json:"applicant_name"`
	InstitutionName  string `json:"institution_name"`
	EducationLevel   string `json:"education_level"`
}

type DashboardSummaryResponse struct {
	Stats              DashboardStats          `json:"stats"`
	Chart              []DashboardChartYear    `json:"chart"`
	RecentApplications []RecentApplicationItem `json:"recent_applications"`
}