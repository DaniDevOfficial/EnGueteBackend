package dates

import "time"

func GetStartAndEndOfWeek(date time.Time) (startOfWeek string, endOfWeek string) {

	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	start := date.AddDate(0, 0, -weekday+1)
	end := start.AddDate(0, 0, 6)

	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}
