package service

import (
	"context"
	"fmt"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/model"
	"internship-telkom/repository"
	"internship-telkom/utils"
	"log"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type ApplicationService interface {
	ApplyInternship(userID uint, input dto.CreateApplicationRequest, files map[string]*multipart.FileHeader) (dto.CreateApplicationResponse, error)
	GetMyApplications(userID uint) ([]dto.ApplicationListResponse, error)
	GetApplicationDetail(id uint, userID uint) (dto.ApplicationDetailResponse, error)
}

type applicationService struct {
	repo repository.ApplicationRepository
	cld  *cloudinary.Cloudinary
	userRepo  repository.UserRepository
	notifRepo repository.NotificationRepository
}

func NewApplicationService(
	repo repository.ApplicationRepository, 
	cld *cloudinary.Cloudinary, userRepo repository.UserRepository, 
	notifRepo repository.NotificationRepository,
	) ApplicationService {
	return &applicationService{repo, cld, userRepo, notifRepo}
}

func (s *applicationService) ApplyInternship(userID uint, input dto.CreateApplicationRequest, files map[string]*multipart.FileHeader) (dto.CreateApplicationResponse, error) {
	// 1. Validasi Dasar: Cek apakah pendaftaran aktif sudah ada
	_, err := s.repo.FindActiveByUserID(userID)
	if err == nil {
		return dto.CreateApplicationResponse{}, apperrors.ErrApplicationStillActive
	}
	// 2. Validasi dokumen wajib berdasarkan jenjang pendidikan
	if err := validateRequiredDocuments(input.EducationLevel, files); err != nil {
		return dto.CreateApplicationResponse{}, err
	}
	// 3. Parse Tanggal
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, input.StartDate)
	if err != nil {
		return dto.CreateApplicationResponse{}, fmt.Errorf("format start_date tidak valid, gunakan YYYY-MM-DD")
	}
	endDate, err := time.Parse(layout, input.EndDate)
	if err != nil {
		return dto.CreateApplicationResponse{}, fmt.Errorf("format end_date tidak valid, gunakan YYYY-MM-DD")
	}
	if !endDate.After(startDate) {
		return dto.CreateApplicationResponse{}, fmt.Errorf("end_date harus setelah start_date")
	}

	app := model.InternshipApplication{
		UserID:              userID,
		EducationLevel:      input.EducationLevel,
		IdentityNumber:      input.IdentityNumber,
		InstitutionName:     input.InstitutionName,
		Major:               input.Major,
		Semester:            input.Semester,
		StartDate:           startDate,
		EndDate:             endDate,
		IntegrityPactAgreed: input.IntegrityPactAgreed,
		Status:              model.StatusPending,
	}

	createdApp, err := s.repo.Save(app)
	if err != nil {
		return dto.CreateApplicationResponse{}, fmt.Errorf("failed to save application: %w", err)
	}
	// 4. Upload File ke Cloudinary & Simpan ke Tabel Documents
	ctx := context.Background()
	var savedDocuments []dto.DocumentResponse

	for docType, fileHeader := range files {
		if fileHeader == nil {
			continue
		}

		file, _ := fileHeader.Open()
		if err != nil {
			return dto.CreateApplicationResponse{}, fmt.Errorf("failed to open file %s: %w", docType, err)
		}
		defer file.Close()

		uploadRes, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
			Folder: "internship_telkom_docs",
			Type: "upload",
		})
		if err != nil {
			return dto.CreateApplicationResponse{}, fmt.Errorf("failed to upload file %s: %w", docType, err)
		}

		doc := model.Document{
			InternshipApplicationID: createdApp.ID,
			DocumentType:            docType,
			FileURL:                 uploadRes.SecureURL,
		}
		if err := s.repo.SaveDocument(doc); err != nil {
			return dto.CreateApplicationResponse{}, fmt.Errorf("failed to save document %s: %w", docType, err)
		}

		savedDocuments = append(savedDocuments, dto.DocumentResponse{
			DocumentType: docType,
			FileURL:      uploadRes.SecureURL,
			UploadedAt:   time.Now().Format("2006-01-02"),
		})
	}
	
	if s.notifRepo != nil && s.userRepo != nil {
		applicant, err := s.userRepo.FindByID(userID)
		applicantName := "Pemohon"
		if err == nil {
			applicantName = applicant.Fullname
		}

		staff, err := s.userRepo.FindAllStaff()
		if err == nil {
			for _, u := range staff {
				notif := model.Notification{
					UserID:                  u.ID,
					InternshipApplicationID: createdApp.ID,
					Description:             fmt.Sprintf("Pengajuan PKL baru dari %s menunggu proses verifikasi.", applicantName),
					IsRead:                  false,
				}
				if _, err := s.notifRepo.Create(notif); err != nil {
					log.Printf("Warning: gagal membuat notifikasi staf untuk aplikasi %d: %v", createdApp.ID, err)
				}
			}
		}
	}

	return dto.CreateApplicationResponse{
		ID:                  createdApp.ID,
		UserID:              createdApp.UserID,
		EducationLevel:      createdApp.EducationLevel,
		IdentityNumber:      createdApp.IdentityNumber,
		InstitutionName:     createdApp.InstitutionName,
		Major:               createdApp.Major,
		SemesterOrClass:     utils.FormatSemesterOrClass(createdApp.EducationLevel, createdApp.Semester),
		StartDate:           createdApp.StartDate.Format("2006-01-02"),
		EndDate:             createdApp.EndDate.Format("2006-01-02"),
		Status:              createdApp.Status,
		IntegrityPactAgreed: createdApp.IntegrityPactAgreed,
		Documents:           savedDocuments,
	}, nil
}

func (s *applicationService) GetMyApplications(userID uint) ([]dto.ApplicationListResponse, error) {
	apps, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}

	var response []dto.ApplicationListResponse
	for _, app := range apps {
		response = append(response, dto.ApplicationListResponse{
			ID:               app.ID,
			RegistrationDate: app.CreatedAt.Format("02 Jan 2006"),
			EducationLevel:   app.EducationLevel,
			Period:           formatPeriod(app.StartDate, app.EndDate),
			Status:           formatStatus(app.Status),
		})
	}
	return response, nil
}

func (s *applicationService) GetApplicationDetail(id uint, userID uint) (dto.ApplicationDetailResponse, error) {
	app, err := s.repo.FindByIDAndUserID(id, userID)
	if err != nil {
		return dto.ApplicationDetailResponse{}, apperrors.ErrApplicationNotFound
	}

	var documents []dto.DocumentResponse
	for _, doc := range app.Documents {
		documents = append(documents, dto.DocumentResponse{
			ID:           doc.ID,
			DocumentType: doc.DocumentType,
			FileURL:      doc.FileURL,
			UploadedAt:   doc.CreatedAt.Format("2006-01-02"),
		})
	}

	var letter *dto.LetterResponse
	if app.Letter != nil {
		letter = &dto.LetterResponse{
			ID:           app.Letter.ID,
			LetterNumber: app.Letter.LetterNumber,
			LetterDate:   app.Letter.LetterDate.Format("02 Jan 2006"),
			FileURL:      app.Letter.FileURL,
			GeneratedAt:  app.Letter.GeneratedAt.Format("02 Jan 2006"),
		}
	}

	adminNotes := ""
	if app.AdminNotes != nil{
		adminNotes = *app.AdminNotes
	}

	attendanceDate := ""
	if app.AttendanceDate != nil {
		attendanceDate = app.AttendanceDate.Format("02 Jan 2006")
	}

	return dto.ApplicationDetailResponse{
		ID:                	 app.ID,
		RegistrationDate:  	 app.CreatedAt.Format("02 Jan 2006"),
		Period:            	 formatPeriod(app.StartDate, app.EndDate),
		Status:              formatStatus(app.Status),
		AdminNotes:       	 adminNotes,
		AttendanceDate:   	 attendanceDate,
		Name:                app.User.Fullname,
		EducationLevel:   	 app.EducationLevel,
		IdentityNumber:      app.IdentityNumber,
		InstitutionName:     app.InstitutionName,
		Major:               app.Major,
		SemesterOrClass:     utils.FormatSemesterOrClass(app.EducationLevel, app.Semester),
		Documents:        	 documents,
		Letter:           	 letter,
	}, nil
}

func validateRequiredDocuments(educationLevel string, files map[string]*multipart.FileHeader) error {
	var required []string
	switch {
	case educationLevel == "SMK":
		required = []string{"recommendation_letter"}
	case utils.IsHigherEducation(educationLevel):
		required = []string{"recommendation_letter", "transcript", "cv", "proposal"}
	default:
		return fmt.Errorf("jenjang pendidikan tidak dikenali: %s", educationLevel)
	}

	for _, doc := range required {
		if files[doc] == nil {
			return fmt.Errorf("dokumen wajib belum diunggah: %s", doc)
		}
	}
	return nil
}

func formatStatus(status string) string{
	switch status {
		case model.StatusPending:
		return "Sedang Diproses"
	case model.StatusApproved:
		return "Diterima"
	case model.StatusRejected:
		return "Ditolak"
	case model.StatusActive:
		return "Aktif"
	case model.StatusDone:
		return "Selesai"
	default:
		return status
	}
}

func formatPeriod(start, end time.Time) string {
	return fmt.Sprintf("%s s.d. %s", start.Format("2 Jan 2006"), end.Format("2 Jan 2006"))
}