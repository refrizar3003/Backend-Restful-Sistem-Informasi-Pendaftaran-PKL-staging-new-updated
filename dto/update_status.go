package dto

type UpdateApplicationStatusRequest struct {
	Status         string `json:"status"`
	AdminNotes     string `json:"admin_notes"`
	AttendanceDate string `json:"attendance_date"`
	UnitID         *uint  `json:"unit_id"`
	MentorID       *uint  `json:"mentor_id"`
}

type UpdateApplicationStatusResponse struct {
	ID         uint   `json:"id"`
	Status     string `json:"status"`
	AdminNotes string `json:"admin_notes,omitempty"`
}