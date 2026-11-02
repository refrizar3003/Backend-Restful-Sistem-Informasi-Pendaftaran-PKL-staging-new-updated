package service

import (
	"fmt"
	"internship-telkom/apperrors"
	"internship-telkom/dto"
	"internship-telkom/infrastructure"
	"internship-telkom/model"
	"internship-telkom/repository"
	"internship-telkom/utils"
	"log"
	"time"
)

type AdminApplicationService interface {
	GetApplications(input dto.AdminApplicationListRequest) (dto.AdminApplicationListResponse, error)
	GetApplicationDetail(id uint) (dto.AdminApplicationDetailResponse, error)
	UpdateApplicationStatus(id uint, input dto.UpdateApplicationStatusRequest, actorID uint) (dto.UpdateApplicationStatusResponse, error)
	
	GetUnits() ([]dto.AdminUnitInfo, error)
	GetMentorsByUnit(unitID uint) ([]dto.AdminMentorInfo, error)

	GetDashboardSummary(year int) (dto.DashboardSummaryResponse, error)
}

type adminApplicationService struct {
	repo repository.ApplicationRepository
	unitRepo   repository.UnitRepository
	mentorRepo repository.MentorRepository
	notifRepo  repository.NotificationRepository
	userRepo   repository.UserRepository
	mailer     infrastructure.Mailer
	letterService LetterService
}

func NewAdminApplicationService(
	repo repository.ApplicationRepository, 
	unitRepo repository.UnitRepository, 
	mentorRepo repository.MentorRepository, 
	notifRepo repository.NotificationRepository, 
	userRepo repository.UserRepository, 
	mailer infrastructure.Mailer, 
	letterService LetterService,
	) AdminApplicationService {
	return &adminApplicationService{repo, unitRepo, mentorRepo, notifRepo, userRepo, mailer, letterService}
}

func (s *adminApplicationService) GetApplications(input dto.AdminApplicationListRequest) (dto.AdminApplicationListResponse, error) {
	page, limit := normalizePagination(input.Page, input.Limit)

	apps, total, err := s.repo.FindAllForAdmin(input.Status, input.Category, input.Search, page, limit)
	if err != nil {
		return dto.AdminApplicationListResponse{}, fmt.Errorf("failed to get applications: %w", err)
	}

	items := make([]dto.AdminApplicationListItem, 0, len(apps))
	for _, app := range apps {
		items = append(items, dto.AdminApplicationListItem{
			ID:               app.ID,
			RegistrationDate: app.CreatedAt.Format("2 - Jan - 2006"),
			ApplicantName:    app.User.Fullname,
			Category:         app.EducationLevel,
			Period:           formatAdminPeriod(app.StartDate, app.EndDate),
			Status:           formatStatus(app.Status),
		})
	}

	return dto.AdminApplicationListResponse{
		Total: int(total),
		Page:  page,
		Limit: limit,
		Items: items,
	}, nil
}

func (s *adminApplicationService) GetApplicationDetail(id uint) (dto.AdminApplicationDetailResponse, error) {
	app, err := s.repo.FindByIDForAdmin(id)
	if err != nil {
		return dto.AdminApplicationDetailResponse{}, apperrors.ErrApplicationNotFound
	}

	documents := make([]dto.DocumentResponse, 0, len(app.Documents))
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
			LetterDate:   app.Letter.LetterDate.Format("2 Jan 2006"),
			FileURL:      app.Letter.FileURL,
			GeneratedAt:  app.Letter.GeneratedAt.Format("2 Jan 2006"),
		}
	}

	var unit *dto.AdminUnitInfo
	if app.Unit != nil {
		unit = &dto.AdminUnitInfo{
			ID:   app.Unit.ID,
			Name: app.Unit.Name,
		}
	}

	var mentor *dto.AdminMentorInfo
	if app.Mentor != nil {
		mentor = &dto.AdminMentorInfo{
			ID:       app.Mentor.ID,
			Name:     app.Mentor.Name,
			Position: app.Mentor.Position,
		}
	}

	adminNotes := ""
	if app.AdminNotes != nil {
		adminNotes = *app.AdminNotes
	}

	attendanceDate := ""
	if app.AttendanceDate != nil {
		attendanceDate = app.AttendanceDate.Format("2 Jan 2006")
	}

	return dto.AdminApplicationDetailResponse{
		ID:               app.ID,
		RegistrationDate: app.CreatedAt.Format("2 Jan 2006"),
		Category:         app.EducationLevel,
		Period:           formatAdminPeriod(app.StartDate, app.EndDate),
		Status:           formatStatus(app.Status),
		AdminNotes:       adminNotes,
		AttendanceDate:   attendanceDate,
		Applicant: dto.AdminApplicantInfo{
			ID:          app.User.ID,
			Fullname:    app.User.Fullname,
			Email:       app.User.Email,
			PhoneNumber: app.User.PhoneNumber,
		},
		Academic: dto.AdminAcademicInfo{
			EducationLevel:  app.EducationLevel,
			InstitutionName: app.InstitutionName,
			Major:           app.Major,
			IdentityNumber:  app.IdentityNumber,
			SemesterOrClass: utils.FormatSemesterOrClass(app.EducationLevel, app.Semester),
		},
		Documents: documents,
		Letter:    letter,
		Unit:      unit,
		Mentor:    mentor,
	}, nil
}

func (s *adminApplicationService) UpdateApplicationStatus(id uint, input dto.UpdateApplicationStatusRequest, actorID uint) (dto.UpdateApplicationStatusResponse, error) {
	if input.Status != model.StatusApproved && input.Status != model.StatusRejected {
		return dto.UpdateApplicationStatusResponse{}, apperrors.ErrInvalidApplicationStatus
	}

	app, err := s.repo.FindByIDForAdmin(id)
	if err != nil {
		return dto.UpdateApplicationStatusResponse{}, apperrors.ErrApplicationNotFound
	}
	if app.Status != model.StatusPending {
		return dto.UpdateApplicationStatusResponse{}, apperrors.ErrApplicationNotPending
	}

	var (
		adminNotes     *string
		unitID         *uint
		mentorID       *uint
		attendanceDate *time.Time
		unitNameForEmail       string
		mentorNameForEmail     string
		attendanceDateForEmail string
	)

	switch input.Status {
		case model.StatusApproved:
			if input.UnitID == nil || input.MentorID == nil || input.AttendanceDate == "" {
				return dto.UpdateApplicationStatusResponse{}, apperrors.ErrApprovedRequiresUnitMentor
		}

		unit, err := s.unitRepo.FindByID(*input.UnitID)
		if err != nil {
			return dto.UpdateApplicationStatusResponse{}, apperrors.ErrApprovedRequiresUnitMentor
		}

		mentor, err := s.mentorRepo.FindByID(*input.MentorID)
		if err != nil {
			return dto.UpdateApplicationStatusResponse{}, apperrors.ErrMentorNotFound
		}
		if mentor.UnitID != *input.UnitID {
			return dto.UpdateApplicationStatusResponse{}, apperrors.ErrMentorNotInUnit
		}

		parsed, err := time.Parse("2006-01-02", input.AttendanceDate)
		if err != nil {
			return dto.UpdateApplicationStatusResponse{}, fmt.Errorf("format attendance_date tidak valid, gunakan YYYY-MM-DD")
		}

		unitID = input.UnitID
		mentorID = input.MentorID
		attendanceDate = &parsed
		unitNameForEmail = unit.Name
		mentorNameForEmail = mentor.Name
		attendanceDateForEmail = parsed.Format("2 Jan 2006")

		if input.AdminNotes != "" {
			adminNotes = &input.AdminNotes
		}

	case model.StatusRejected:
		if input.AdminNotes == "" {
			return dto.UpdateApplicationStatusResponse{}, apperrors.ErrRejectedRequiresNotes
		}
		adminNotes = &input.AdminNotes
	}

	if err := s.repo.UpdateStatus(id, input.Status, adminNotes, unitID, mentorID, attendanceDate); err != nil {
		return dto.UpdateApplicationStatusResponse{}, fmt.Errorf("failed to update application status: %w", err)
	}

	if input.Status == model.StatusRejected && s.letterService != nil {
		go func(appID uint) {
			_, err := s.letterService.GenerateLetter(appID, dto.GenerateLetterRequest{})
			if err != nil {
				log.Printf("Warning: gagal membuat surat penolakan untuk application %d: %v", appID, err)
			}
		} (id)
	}

	if s.notifRepo != nil {
		var description string
    	switch input.Status {
    	case model.StatusApproved:
        	description = "Selamat! Pengajuan PKL kamu telah diterima. Silakan buka menu \"Lihat Surat Penerimaan\" pada halaman ini untuk mengunduh surat resmi kamu."
    	case model.StatusRejected:
        	description = fmt.Sprintf("Mohon maaf, pengajuan PKL kamu ditolak. Alasan: %s", input.AdminNotes)
		}
		notif := model.Notification{
			UserID:                  app.UserID,
			InternshipApplicationID: app.ID,
			Description:             description,
			IsRead:                  false,
    	}
    	if _, err := s.notifRepo.Create(notif); err != nil {
			log.Printf("Warning: gagal membuat notifikasi untuk user %d (application %d): %v", app.UserID, app.ID, err)
		}

		if s.mailer != nil {
			go func(status, email, fullname, notes, unitName, mentorName, attDate string) {
				var sendErr error
				switch status {
				case model.StatusApproved:
					sendErr = s.mailer.SendApplicationAcceptedEmail(email, fullname, unitName, mentorName, attDate)
				case model.StatusRejected:
					sendErr = s.mailer.SendApplicationRejectedEmail(email, fullname, notes)
				}
				if sendErr != nil {
					log.Printf("Warning: gagal mengirim email status PKL ke %s (application %d): %v", email, app.ID, sendErr)
				}
			}(input.Status, app.User.Email, app.User.Fullname, input.AdminNotes, unitNameForEmail, mentorNameForEmail, attendanceDateForEmail)
		}

		actorName := "Admin"
		if actorID != 0 && s.userRepo != nil {
			if actor, err := s.userRepo.FindByID(actorID); err == nil {
				actorName = actor.Fullname
			}
		}

		var staffDescription string
		switch input.Status {
		case model.StatusApproved:
			staffDescription = fmt.Sprintf("Pengajuan PKL %s telah diterima dan surat penerimaan telah dikirim oleh %s.", app.User.Fullname, actorName)
		case model.StatusRejected:
			staffDescription = fmt.Sprintf("Pengajuan PKL %s telah ditolak oleh %s.", app.User.Fullname, actorName)
		}
		
		if s.userRepo != nil {
			staff, err := s.userRepo.FindAllStaff()
			if err == nil {
				for _, u := range staff {
					staffNotif := model.Notification{
						UserID:                  u.ID,
                    	InternshipApplicationID: app.ID,
                    	Description:             staffDescription,
                    	IsRead:                  false,
					}
					if _, err := s.notifRepo.Create(staffNotif); err != nil {
						log.Printf("Warning: gagal membuat notifikasi staf untuk aplikasi %d: %v", app.ID, err)
					}
				}
			}
		}
	}
	
	notes := ""
	if adminNotes != nil {
		notes = *adminNotes
	}

	return dto.UpdateApplicationStatusResponse{
		ID:         id,
		Status:     formatStatus(input.Status),
		AdminNotes: notes,
	}, nil
}

func (s *adminApplicationService) GetUnits() ([]dto.AdminUnitInfo, error) {
	units, err := s.unitRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get units: %w", err)
	}

	result := make([]dto.AdminUnitInfo, 0, len(units))
	for _, u := range units {
		result = append(result, dto.AdminUnitInfo{ID: u.ID, Name: u.Name})
	}
	return result, nil
}

func (s *adminApplicationService) GetMentorsByUnit(unitID uint) ([]dto.AdminMentorInfo, error) {
	if _, err := s.unitRepo.FindByID(unitID); err != nil {
		return nil, apperrors.ErrUnitNotFound
	}

	mentors, err := s.mentorRepo.FindByUnitID(unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentors: %w", err)
	}

	result := make([]dto.AdminMentorInfo, 0, len(mentors))
	for _, m := range mentors {
		result = append(result, dto.AdminMentorInfo{
			ID:       m.ID,
			Name:     m.Name,
			Position: m.Position,
		})
	}
	return result, nil
}

func (s *adminApplicationService) GetDashboardSummary(yearsBack int) (dto.DashboardSummaryResponse, error) {
	if yearsBack <= 0 {
		yearsBack = 5
	}
	
	pendaftarSMK, err := s.repo.CountByStatusAndEducationLevel(model.StatusPending, "SMK")
	if err != nil {
		return dto.DashboardSummaryResponse{}, err
	}
	pendaftarD3, err := s.repo.CountByStatusAndEducationLevel(model.StatusPending, "D3")
	if err != nil {
		return dto.DashboardSummaryResponse{}, err
	}
	pendaftarD4, err := s.repo.CountByStatusAndEducationLevel(model.StatusPending, "D4")
	if err != nil {
		return dto.DashboardSummaryResponse{}, err
	}
	pendaftarS1, err := s.repo.CountByStatusAndEducationLevel(model.StatusPending, "S1")
	if err != nil {
		return dto.DashboardSummaryResponse{}, err
	}

	currentYear := time.Now().Year()
	chart := make([]dto.DashboardChartYear, 0, yearsBack)
	for i := yearsBack - 1; i >= 0; i-- {
		year := currentYear - i

		total, err := s.repo.CountByYear(year)
		if err != nil {
			return dto.DashboardSummaryResponse{}, err
		}
		diterima, err := s.repo.CountByStatusAndYear(model.StatusApproved, year)
		if err != nil {
			return dto.DashboardSummaryResponse{}, err
		}
		ditolak, err := s.repo.CountByStatusAndYear(model.StatusRejected, year)
		if err != nil {
			return dto.DashboardSummaryResponse{}, err
		}

		chart = append(chart, dto.DashboardChartYear{
			Year:     year,
			Total:    total,
			Diterima: diterima,
			Ditolak:  ditolak,
		})
	}

	recentApps, err := s.repo.FindRecentApplications(5)
	if err != nil {
		return dto.DashboardSummaryResponse{}, err
	}

	recentItems := make([]dto.RecentApplicationItem, 0, len(recentApps))
	for _, app := range recentApps {
		recentItems = append(recentItems, dto.RecentApplicationItem{
			ID:               app.ID,
			RegistrationDate: app.CreatedAt.Format("2 - Jan - 2006"),
			ApplicantName:    app.User.Fullname,
			InstitutionName:  app.InstitutionName,
			EducationLevel:   app.EducationLevel,
		})
	}

	return dto.DashboardSummaryResponse{
		Stats: dto.DashboardStats{
			PendaftarSMK: pendaftarSMK,
			PendaftarD3:  pendaftarD3,
			PendaftarD4:  pendaftarD4,
			PendaftarS1:  pendaftarS1,
		},
		Chart:              chart,
		RecentApplications: recentItems,
	}, nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func formatAdminPeriod(start, end time.Time) string {
	return fmt.Sprintf("%s s.d. %s", utils.FormatIndonesianDate(start), utils.FormatIndonesianDate(end))
}