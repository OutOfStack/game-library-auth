package types

import (
	"database/sql"
)

func NullTimeString(time sql.NullTime) string {
	if time.Valid {
		return time.Time.String()
	} else {
		return ""
	}
}
