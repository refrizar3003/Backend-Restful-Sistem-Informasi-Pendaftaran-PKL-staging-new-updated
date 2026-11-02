package utils

import (
	"fmt"
	"time"
)

var indonesianMonths = [...]string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

func FormatIndonesianDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), indonesianMonths[t.Month()-1], t.Year())
}