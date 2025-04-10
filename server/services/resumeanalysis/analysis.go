package resumeanalysis

type ResumeSkills struct {
	HardSkills []string
	SoftSkills []string
}

type VacancySkills struct {
	HardSkills []string
	SoftSkills []string
}

func AnalizResumeSkills(resskills ResumeSkills, vacskills VacancySkills) float64 {
	softResumeSkills := resskills.SoftSkills
	hardResumeSkills := resskills.HardSkills
	softVacancySkills := vacskills.SoftSkills
	hardVacancySkills := vacskills.HardSkills
	count_soft_resskills := 0
	count_hard_resskills := 0
	for _, soft_resskill := range softResumeSkills {
		for _, soft_vacskill := range softVacancySkills {
			if soft_resskill == soft_vacskill {
				count_soft_resskills++
				break
			}
		}
	}
	for _, hard_resskill := range hardResumeSkills {
		for _, hard_vacskill := range hardVacancySkills {
			if hard_resskill == hard_vacskill {
				count_hard_resskills++
				break
			}
		}
	}
	softScore := float64(count_soft_resskills) / float64(len(softVacancySkills)) * 100
	hardScore := float64(count_hard_resskills) / float64(len(hardVacancySkills)) * 100

	if len(hardVacancySkills) == 0 {
		return softScore
	} else if len(softVacancySkills) == 0 {
		return hardScore
	} else {
		return (softScore + hardScore) / 2
	}
}
