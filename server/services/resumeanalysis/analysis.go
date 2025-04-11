package resumeanalysis

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

func AnalizResumeSkills(resskills models.ResumeSkills, vacskills models.VacancySkills) (models.FinalSkills, error) {
	var finalSkills models.FinalSkills

	countSoft := 0
	countHard := 0

	for _, softVacSkill := range vacskills.SoftSkills {
		found := false
		for _, softResSkill := range resskills.SoftSkills {
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

	for _, hardVacSkill := range vacskills.HardSkills {
		found := false
		for _, hardResSkill := range resskills.HardSkills {
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
	if len(vacskills.SoftSkills) > 0 {
		softScore = float64(countSoft) / float64(len(vacskills.SoftSkills)) * 100
	}
	if len(vacskills.HardSkills) > 0 {
		hardScore = float64(countHard) / float64(len(vacskills.HardSkills)) * 100
	}

	switch {
	case len(vacskills.HardSkills) == 0:
		finalSkills.Percent = int(softScore)
	case len(vacskills.SoftSkills) == 0:
		finalSkills.Percent = int(hardScore)
	default:
		finalSkills.Percent = int((softScore + hardScore) / 2)
	}

	return finalSkills, nil
}
