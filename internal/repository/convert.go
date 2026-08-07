package repository

import "github.com/jackc/pgx/v5/pgtype"

func addDays(d pgtype.Date, days int) pgtype.Date {
	return pgtype.Date{Time: d.Time.AddDate(0, 0, days), Valid: true}
}
