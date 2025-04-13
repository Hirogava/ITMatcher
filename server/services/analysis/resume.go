package analysis

import (
	"strings"

	"gaspr/models"

	"github.com/agnivade/levenshtein"
)

const maxDistance = 2

func skillMatch(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return true
	}
	distance := levenshtein.ComputeDistance(a, b)
	return distance <= maxDistance
}

func AnalyseResumeSkills(resSkills models.ResumeSkills, vacSkills models.VacancySkills) (models.FinalSkills, error) {
	var finalSkills models.FinalSkills

	countSoft := 0
	countHard := 0

	for _, softVacSkill := range vacSkills.SoftSkills {
		found := false
		for _, softResSkill := range resSkills.SoftSkills {
			if skillMatch(softResSkill.SkillName, softVacSkill.SkillName) {
				finalSkills.CoincidenceSoft = append(finalSkills.CoincidenceSoft, softVacSkill)
				countSoft++
				found = true
				break
			}
		}
		if !found {
			finalSkills.MismatchSoft = append(finalSkills.MismatchSoft, softVacSkill)
		}
	}

	for _, hardVacSkill := range vacSkills.HardSkills {
		found := false
		for _, hardResSkill := range resSkills.HardSkills {
			if skillMatch(hardResSkill.SkillName, hardVacSkill.SkillName) {
				finalSkills.CoincidenceHard = append(finalSkills.CoincidenceHard, hardVacSkill)
				countHard++
				found = true
				break
			}
		}
		if !found {
			finalSkills.MismatchHard = append(finalSkills.MismatchHard, hardVacSkill)
		}
	}

	var softScore, hardScore float64
	if len(vacSkills.SoftSkills) > 0 {
		softScore = float64(countSoft) / float64(len(vacSkills.SoftSkills)) * 100
	}
	if len(vacSkills.HardSkills) > 0 {
		hardScore = float64(countHard) / float64(len(vacSkills.HardSkills)) * 100
	}

	switch {
	case len(vacSkills.HardSkills) == 0:
		finalSkills.Percent = int(softScore)
	case len(vacSkills.SoftSkills) == 0:
		finalSkills.Percent = int(hardScore)
	default:
		finalSkills.Percent = int((softScore + hardScore) / 2)
	}

	return finalSkills, nil
}
