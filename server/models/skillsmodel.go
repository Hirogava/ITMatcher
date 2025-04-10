package models

type ResumeSkills struct {
	HardSkills []string
	SoftSkills []string
}

type VacancySkills struct {
	HardSkills []string
	SoftSkills []string
}

type FinalSkills struct {
	Percent int
	CoincidenceHard []string
	CoincidenceSoft []string
	MismatchHard []string
	MismatchSoft []string
}