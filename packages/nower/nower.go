package nower

import "time"

var DefaultNower = TimeNower{}

type TimeNower struct{}

func (n TimeNower) Now() time.Time {
	return time.Now()
}

// returns last month as string in format yyyy-mm
// e.g. if its 2025-12-13, return 2025-11 (M - 1)
func GetLastMonthYYYYMM(nower Nower) string {
	now := nower.Now()
	lastMonthTime := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())

	return lastMonthTime.Format("2006-01")
}

type Nower interface {
	Now() time.Time
}
