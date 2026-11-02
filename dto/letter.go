package dto

type GenerateLetterRequest struct {
	LetterNumber    string `json:"letter_number"  validate:"required"`
	LetterDate      string `json:"letter_date"    validate:"required"`
	ReferenceNumber string `json:"reference_number"`
	RecipientName   string `json:"recipient_name" validate:"required"`
	SignerName      string `json:"signer_name"    validate:"required"`
	SignerTitle     string `json:"signer_title"   validate:"required"`
	SignerNik       string `json:"signer_nik"     validate:"required"`
}

type GenerateLetterResponse struct {
	ID           uint   `json:"id"`
	LetterNumber string `json:"letter_number"`
	LetterDate   string `json:"letter_date"`
	FileURL      string `json:"file_url"`
	GeneratedAt  string `json:"generated_at"`
}

type GenerateRejectionLetterRequest struct {
	RejectionReason string `json:"rejection_reason" validate:"required"`

	LetterNumber    string `json:"letter_number"`
	LetterDate      string `json:"letter_date"`
	ReferenceNumber string `json:"reference_number"`
	RecipientName   string `json:"recipient_name"`
	SignerName      string `json:"signer_name"`
	SignerTitle     string `json:"signer_title"`
	SignerNik       string `json:"signer_nik"`
}