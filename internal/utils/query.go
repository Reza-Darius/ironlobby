package utils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Text converts string to pgtype Text
func Text(v string) pgtype.Text {
	if len(v) == 0 {
		return pgtype.Text{
			String: "",
			Valid:  false,
		}
	}
	return pgtype.Text{
		String: v,
		Valid:  true,
	}
}

func Time(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{
			Time:             time.Time{},
			InfinityModifier: 0,
			Valid:            false,
		}
	}
	return pgtype.Timestamptz{
		Time:             *v,
		InfinityModifier: 0,
		Valid:            true,
	}
}
