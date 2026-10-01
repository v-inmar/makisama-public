package date_utils

import (
	"database/sql"
	"time"
)

// returns midnight of the seventh day from now. utc timezone
func GetMidnightOfSeventhDay() time.Time {
	future := time.Now().AddDate(0, 0, 7)
	midnight := time.Date(future.Year(), future.Month(), future.Day(), 0, 0, 0, 0, time.UTC)
	return midnight
}

func FormatNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.RFC3339)
}

func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
