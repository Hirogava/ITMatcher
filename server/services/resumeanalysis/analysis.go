package resumeanalysis

import (
	"gaspr/models"
)

func AnalizResumeSkills(resskills models.ResumeSkills, vacskills models.VacancySkills) models.FinalSkills {
	var finalSkills models.FinalSkills
	softResumeSkills := resskills.SoftSkills
	hardResumeSkills := resskills.HardSkills
	softVacancySkills := vacskills.SoftSkills
	hardVacancySkills := vacskills.HardSkills
	count_soft_resskills := 0
	count_hard_resskills := 0
	for _, soft_resskill := range softResumeSkills {
		for _, soft_vacskill := range softVacancySkills {
			if soft_resskill == soft_vacskill {
				finalSkills.CoincidenceSoft = append(finalSkills.CoincidenceSoft, soft_vacskill)
				count_soft_resskills++
				break
			}
			finalSkills.MismatchSoft = append(finalSkills.MismatchSoft, soft_vacskill)
		}
	}
	for _, hard_resskill := range hardResumeSkills {
		for _, hard_vacskill := range hardVacancySkills {
			if hard_resskill == hard_vacskill {
				finalSkills.CoincidenceHard = append(finalSkills.CoincidenceHard, hard_vacskill)
				count_hard_resskills++
				break
			}
			finalSkills.MismatchHard = append(finalSkills.MismatchHard, hard_vacskill)
		}
	}
	softScore := float64(count_soft_resskills) / float64(len(softVacancySkills)) * 100
	hardScore := float64(count_hard_resskills) / float64(len(hardVacancySkills)) * 100

	if len(hardVacancySkills) == 0 {
		finalSkills.Percent = int(softScore)
		return finalSkills
	} else if len(softVacancySkills) == 0 {
		finalSkills.Percent = int(hardScore)
		return finalSkills
	} else {
		finalSkills.Percent = int((softScore + hardScore) / 2)
		return finalSkills
	}
}
