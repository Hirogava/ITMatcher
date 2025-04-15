package models

type ResumeSoftSkill struct {
	Id        int
	SkillName string
}

type ResumeHardSkill struct {
	Id        int
	SkillName string
}

type VacancySoftSkill struct {
	Id        int
	SkillName string
}

type VacancyHardSkill struct {
	Id        int
	SkillName string
}

type ResumeSkills struct {
	HardSkills []ResumeHardSkill
	SoftSkills []ResumeSoftSkill
}

type VacancySkills struct {
	HardSkills []VacancyHardSkill
	SoftSkills []VacancySoftSkill
}

type FinalSkills struct {
	Percent         int
	CoincidenceHard []VacancyHardSkill
	CoincidenceSoft []VacancySoftSkill
	MismatchHard    []VacancyHardSkill
	MismatchSoft    []VacancySoftSkill
}

type AnalyzedSkills struct {
	Mismatch    []string `json:"mismatch"`
	Coincidence []string `json:"coincide"`
}

type VacancyMatchResult struct {
	VacancyId   int         `json:"vacancy_id"`
	MatchRate   int         `json:"match_rate"`
	FinalSkills FinalSkills `json:"final_skills"`
}
