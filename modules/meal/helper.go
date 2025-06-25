package meal

import "sort"

func MergeAndSortParticipants(withPreference, withoutPreference []MealPreferences) []MealPreferences {
	allParticipants := append(withPreference, withoutPreference...)

	sort.Slice(allParticipants, func(i, j int) bool {
		if allParticipants[i].Preference == "undecided" && allParticipants[j].Preference != "undecided" {
			return false
		}
		if allParticipants[j].Preference == "undecided" && allParticipants[i].Preference != "undecided" {
			return true
		}

		if allParticipants[i].Preference != allParticipants[j].Preference {
			return allParticipants[i].Preference < allParticipants[j].Preference
		}

		return allParticipants[i].Username < allParticipants[j].Username
	})

	return allParticipants
}
