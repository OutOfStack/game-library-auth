package types

import "database/sql"

// NullTimeString represents sql.NullTime as string
func NullTimeString(time sql.NullTime) string {
	if time.Valid {
		return time.Time.String()
	}
	return ""
}
