package service

import (
	"bytes"
	"context"
	"fmt"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/model"
	"internship-telkom/repository"
	"internship-telkom/utils"
	"net/http"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/jung-kurt/gofpdf"
)

type LetterService interface {
	GenerateLetter(applicationID uint, input dto.GenerateLetterRequest) (dto.GenerateLetterResponse, error)
}

type letterService struct {
	appRepo    repository.ApplicationRepository
	letterRepo repository.LetterRepository
	cld        *cloudinary.Cloudinary
}

func NewLetterService(
	appRepo repository.ApplicationRepository,
	letterRepo repository.LetterRepository,
	cld *cloudinary.Cloudinary,
) LetterService {
	return &letterService{appRepo, letterRepo, cld}
}

func (s *letterService) GenerateLetter(applicationID uint, input dto.GenerateLetterRequest) (dto.GenerateLetterResponse, error) {
	app, err := s.appRepo.FindByIDForAdmin(applicationID)
	if err != nil {
		return dto.GenerateLetterResponse{}, apperrors.ErrApplicationNotFound
	}

	if app.Status != model.StatusApproved && app.Status != model.StatusRejected {
		return dto.GenerateLetterResponse{}, apperrors.ErrLetterRequiresFinalStatus
	}

	existing, err := s.letterRepo.FindByApplicationID(applicationID)
	if err == nil && existing.ID != 0 {
		return dto.GenerateLetterResponse{}, apperrors.ErrLetterAlreadyGenerated
	}

	if app.Status == model.StatusApproved {
		if input.RecipientName == "" || input.SignerName == "" || input.SignerTitle == "" ||
			input.SignerNik == "" || input.LetterNumber == "" || input.LetterDate == "" {
			return dto.GenerateLetterResponse{}, apperrors.ErrAcceptanceLetterMissingFields
		}
	}

	letterDate := time.Now()
	if input.LetterDate != "" {
		parsed, err := time.Parse("2006-01-02", input.LetterDate)
		if err != nil {
			return dto.GenerateLetterResponse{}, fmt.Errorf("format letter_date tidak valid, gunakan YYYY-MM-DD")
		}
		letterDate = parsed
	}

	letterNumber := input.LetterNumber
	recipientName := input.RecipientName
	signerName := input.SignerName
	signerTitle := input.SignerTitle
	signerNik := input.SignerNik

	if app.Status == model.StatusRejected {
		if letterNumber == "" {
			letterNumber = fmt.Sprintf("Tel/PS 000/R2 %d/%d", applicationID, letterDate.Year())
		}
		if recipientName == "" {
			recipientName = fmt.Sprintf("Ketua Jurusan %s %s", app.Major, app.InstitutionName)
		}
		if signerName == "" {
			signerName = getEnvOr("DEFAULT_SIGNER_NAME", "BUDI")
		}
		if signerTitle == "" {
			signerTitle = getEnvOr("DEFAULT_SIGNER_TITLE", "MANAGER SHARED SERVICE LAMPUNG")
		}
		if signerNik == "" {
			signerNik = getEnvOr("DEFAULT_SIGNER_NIK", "710116")
		}
	}

	filledInput := dto.GenerateLetterRequest{
		LetterNumber:    letterNumber,
		LetterDate:      input.LetterDate,
		ReferenceNumber: input.ReferenceNumber,
		RecipientName:   recipientName,
		SignerName:      signerName,
		SignerTitle:     signerTitle,
		SignerNik:       signerNik,
	}

	pdfBytes, err := renderLetterPDF(app, filledInput, letterDate)
	if err != nil {
		return dto.GenerateLetterResponse{}, fmt.Errorf("gagal membuat PDF: %w", err)
	}

	ctx := context.Background()
	filename := fmt.Sprintf("surat_penerimaan_%d_%s", applicationID, time.Now().Format("20060102150405"))

	uploadRes, err := s.cld.Upload.Upload(ctx, bytes.NewReader(pdfBytes), uploader.UploadParams{
		Folder:       "internship_letters",
		PublicID:     filename,
		ResourceType: "raw",
	})
	if err != nil {
		return dto.GenerateLetterResponse{}, fmt.Errorf("gagal upload surat ke storage: %w", err)
	}

	letter := model.Letter{
		InternshipApplicationID: applicationID,
		LetterNumber:            input.LetterNumber,
		LetterDate:              letterDate,
		FileURL:                 uploadRes.SecureURL,
	}

	savedLetter, err := s.letterRepo.Save(letter)
	if err != nil {
		return dto.GenerateLetterResponse{}, fmt.Errorf("gagal menyimpan data surat: %w", err)
	}

	return dto.GenerateLetterResponse{
		ID:           savedLetter.ID,
		LetterNumber: savedLetter.LetterNumber,
		LetterDate:   savedLetter.LetterDate.Format("02 January 2006"),
		FileURL:      savedLetter.FileURL,
		GeneratedAt:  savedLetter.GeneratedAt.Format("02 January 2006"),
	}, nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func renderLetterPDF(app model.InternshipApplication, input dto.GenerateLetterRequest, letterDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(25, 20, 20)
	pdf.SetAutoPageBreak(true, 30)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	contentWidth := pageWidth - 45

	// ── HEADER ───────────────────────────────────────────────────────────────
	logoLoaded := false

	logoURL := os.Getenv("COMPANY_LOGO_URL")
	if logoURL != "" {
		resp, err := http.Get(logoURL)
		if err == nil {
			defer resp.Body.Close()
			pdf.RegisterImageReader("logo", "PNG", resp.Body)
			if pdf.Error() == nil {
				logoLoaded = true
			}
		}
	}

	if logoLoaded {
		pdf.ImageOptions("logo", pageWidth-50, 15, 30, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.SetXY(25, 20)

	} else {
		pdf.SetFont("Arial", "B", 18)
		pdf.SetTextColor(220, 0, 0)
		pdf.CellFormat(contentWidth-40, 8, "Telkom", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(contentWidth-40, 5, "Indonesia", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(contentWidth-40, 4, "the world in your hand", "", 1, "L", false, 0, "")
	}

	currentY := pdf.GetY()
	if currentY < 40 {
		currentY = 40
	}
	pdf.SetY(currentY + 2)
	pdf.SetDrawColor(220, 0, 0)
	pdf.SetLineWidth(0.8)
	pdf.Line(25, pdf.GetY(), pageWidth-20, pdf.GetY())
	pdf.Ln(6)

	// ── NOMOR & TANGGAL ───────────────────────────────────────────────────────
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
 
	labelW := float64(38)
	colonW := float64(5)
	addressLineH := float64(5)
 
	pdf.CellFormat(labelW, addressLineH, "Nomor", "", 0, "L", false, 0, "")
	pdf.CellFormat(colonW, addressLineH, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, addressLineH, input.LetterNumber, "", 1, "L", false, 0, "")
	pdf.Ln(4)
 
	pdf.CellFormat(0, addressLineH, fmt.Sprintf("Bandar Lampung, %s", utils.FormatIndonesianDate(letterDate)), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// ── TUJUAN SURAT ──────────────────────────────────────────────────────────
	pdf.CellFormat(0, addressLineH, "Kepada Yth.", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, addressLineH, input.RecipientName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, addressLineH, "di -", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, addressLineH, "Tempat", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.CellFormat(labelW, addressLineH, "Lampiran", "", 0, "L", false, 0, "")
	pdf.CellFormat(colonW, addressLineH, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, addressLineH, "-", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	if app.Status == model.StatusApproved {
		// ── PERIHAL ───────────────────────────────────────────────────────────────
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(labelW, addressLineH, "Perihal", "", 0, "L", false, 0, "")
		pdf.CellFormat(colonW, addressLineH, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, addressLineH, "Penerimaan Mahasiswa Praktik Kerja Lapangan Witel Lampung", "", 1, "L", false, 0, "")
		pdf.Ln(4)

		// ── BODY PARAGRAPH 1 ─────────────────────────────────────────────────────
		pdf.SetFont("Arial", "", 10)

		lamaMinggu := int(app.EndDate.Sub(app.StartDate).Hours()/24/7) + 1
		mulai := utils.FormatIndonesianDate(app.StartDate)
		selesai := utils.FormatIndonesianDate(app.EndDate)

		unitName := "-"
		if app.Unit != nil {
			unitName = app.Unit.Name
		}

		refNumber := input.ReferenceNumber
		if refNumber == "" {
			refNumber = "..."
		}

		body1 := fmt.Sprintf(
			"      Menunjuk surat Saudara Nomor: %s perihal Permohonan Praktek Kerja Lapangan, "+
				"pada prinsipnya Kantor Wilayah Telkom Lampung menyetujui dan menerima %s atas nama "+
				"%s dari %s untuk melaksanakan Praktik Kerja Lapangan selama %d Minggu terhitung "+
				"mulai %s sampai dengan %s.",
			refNumber,
			educationLabel(app.EducationLevel),
			app.User.Fullname,
			app.InstitutionName,
			lamaMinggu,
			mulai,
			selesai,
		)
		pdf.MultiCell(contentWidth, 6, body1, "", "J", false)
		pdf.Ln(4)

		// ── TABEL PESERTA ─────────────────────────────────────────────────────────
		colW := []float64{8, 40, 28, 35, 28, 36}
		headers := []string{"No.", "Nama", "NPM/NIS", "Jurusan", "Unit", "Pembimbing"}
		rowH := float64(7)
		
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowH+rowH > pageHeight-30 {
			pdf.AddPage()
		}

		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(230, 230, 230)
		for i, h := range headers {
			pdf.CellFormat(colW[i], rowH, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		mentorName := "-"
		if app.Mentor != nil {
			mentorName = app.Mentor.Name
		}

		pdf.SetFont("Arial", "", 8)
		rowData := []string{
			"1.",
			app.User.Fullname,
			app.IdentityNumber,
			app.Major,
			unitName,
			mentorName,
		}

		lineH := float64(5)
		maxLines := 1
		for i, d := range rowData {
			lines := pdf.SplitLines([]byte(d), colW[i]-2)
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}
		cellHeight := float64(maxLines) * lineH
		if cellHeight < rowH {
			cellHeight = rowH
		}

		startX := float64(25)
		startY := pdf.GetY()

		for i, d := range rowData {
			pdf.SetXY(startX, startY)
			pdf.CellFormat(colW[i], cellHeight, "", "1", 0, "C", false, 0, "")
			lines := pdf.SplitLines([]byte(d), colW[i]-2)
			textY := startY + (cellHeight-float64(len(lines))*lineH)/2
			for _, line := range lines {
				textX := startX + (colW[i]-pdf.GetStringWidth(string(line)))/2
				pdf.Text(textX, textY+lineH*0.7, string(line))
				textY += lineH
			}
			startX += colW[i]
		}
		pdf.SetY(startY + cellHeight)
		pdf.Ln(4)

		// ── BODY PARAGRAPH 2 ─────────────────────────────────────────────────────
		pdf.SetFont("Arial", "", 10)

		tanggalHadir := utils.FormatIndonesianDate(app.StartDate)

		body2 := fmt.Sprintf(
			"      Kepada peserta Praktik Kerja Lapangan tersebut di atas kami harapkan kehadirannya "+
				"untuk menemui Manager Shared Service Witel Lampung yang beralamat di Kantor Wilayah "+
				"Telkom Lampung, Jalan Majapahit No. 14 Bandar Lampung 35118 pada tanggal %s.",
			tanggalHadir,
		)
		pdf.MultiCell(contentWidth, 6, body2, "", "J", false)
		pdf.Ln(2)

		pdf.MultiCell(contentWidth, 6,
			"      Demikian kami sampaikan, atas perhatiannya diucapkan terima kasih.",
			"", "J", false)
		pdf.Ln(8)
	} else {
		// ── Surat Penolakan ───────────────────────────────────────────────────
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(labelW, addressLineH, "Perihal", "", 0, "L", false, 0, "")
		pdf.CellFormat(colonW, addressLineH, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, addressLineH, "Penolakan Permohonan Praktik Kerja Lapangan Witel Lampung", "", 1, "L", false, 0, "")
		pdf.Ln(4)

		pdf.SetFont("Arial", "", 10)
		refNumber := input.ReferenceNumber
		if refNumber == "" {
			refNumber = "..."
		}
		body1 := fmt.Sprintf(
			"		Menunjuk surat Saudara Nomor: %s perihal Permohonan Praktik Kerja Lapangan, "+
				"dengan ini kami sampaikan bahwa setelah dilakukan proses seleksi, Kantor Wilayah "+
				"Telkom Lampung belum dapat memenuhi permohonan %s atas nama %s dari %s "+
				"untuk melaksanakan Praktik Kerja Lapangan pada periode yang diajukan.",
			refNumber, educationLabel(app.EducationLevel), app.User.Fullname, app.InstitutionName,
		)
		pdf.MultiCell(contentWidth, 6, body1, "", "J", false)
		pdf.Ln(4)

		reason := "persyaratan administrasi belum terpenuhi"
		if app.AdminNotes != nil && *app.AdminNotes != "" {
			reason = *app.AdminNotes
		}
		body2 := fmt.Sprintf("      Adapun alasan penolakan permohonan tersebut adalah sebagai berikut: %s.", reason)
		pdf.MultiCell(contentWidth, 6, body2, "", "J", false)
		pdf.Ln(2)

		pdf.MultiCell(contentWidth, 6,
			"      	Kami mengucapkan terima kasih atas minat Saudara/i untuk melaksanakan Praktik "+
				"Kerja Lapangan di Kantor Wilayah Telkom Lampung, dan berharap dapat bekerja sama "+
				"pada kesempatan lain. Demikian kami sampaikan, atas perhatiannya "+
				"diucapkan terima kasih.",
			"", "J", false)
		pdf.Ln(8)
	}

	// ── TANDA TANGAN ──────────────────────────────────────────────────────────
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, "Hormat Kami,", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, input.SignerTitle, "", 1, "L", false, 0, "")

	// ── GAMBAR TANDA TANGAN + CAP ─────────────────────────────────────────────
	signatureY := pdf.GetY() - 3
	signatureLoaded := false

	signatureURL := os.Getenv("SIGNATURE_STAMP_URL")
	if signatureURL != "" {
		resp, err := http.Get(signatureURL)
		if err == nil {
			defer resp.Body.Close()
			pdf.RegisterImageReader("signature", "PNG", resp.Body)
			if pdf.Error() == nil{
				pdf.ImageOptions("signature", 25, signatureY, 45, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
				signatureLoaded = true
			}
		}
	}

	if signatureLoaded {
		pdf.SetY(signatureY + 22)
	} else {
		pdf.SetY(signatureY + 16)
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, input.SignerName, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, "NIK. "+input.SignerNik, "", 1, "L", false, 0, "")
	pdf.Ln(8)

	// ── FOOTER ────────────────────────────────────────────────────────────────
	pdf.SetDrawColor(220, 0, 0)
	pdf.SetLineWidth(0.8)
	pdf.Line(25, pdf.GetY(), pageWidth-20, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(0, 5, "PT TELEKOMUNIKASI INDONESIA, Tbk", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 4, "HC Witel Lampung", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, "Jl. Majapahit No. 14 Telp. (0721) 264945", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, "Bandar Lampung 35118", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("gagal export PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func educationLabel(educationLevel string) string {
	if utils.IsHigherEducation(educationLevel) {
		return "mahasiswa"
	}
	if educationLevel == "SMK" {
		return "siswa"
	}
	return "peserta"
}