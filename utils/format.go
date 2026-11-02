package utils

import "fmt"

func IsHigherEducation(educationLevel string) bool {
	switch educationLevel {
	case "S1", "D3", "D4":
		return true
	default:
		return false
	}
}

// FormatSemesterOrClass memformat angka semester/kelas menjadi label
// yang kontekstual berdasarkan jenjang pendidikan.
//
// Contoh:
//   - EducationLevel "Siswa",     Semester 11 → "Kelas 11"
//   - EducationLevel "Mahasiswa", Semester 6  → "Semester 6"
func FormatSemesterOrClass(educationLevel string, semester int) string {
	if IsHigherEducation(educationLevel) {
		return fmt.Sprintf("Semester %d", semester)
	}
	if educationLevel == "SMK" {
		return fmt.Sprintf("Kelas %d", semester)
	}
	return fmt.Sprintf("%d", semester)
}
 