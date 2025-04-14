package models

type Vacancy struct {
	Id          int 				`json:"id"`
	Name        string				`json:"name"`
	HardSkills  []VacancyHardSkill  `json:"hard_skills"`
	SoftSkills  []VacancySoftSkill  `json:"soft_skills"`
	VacancyText string			    `json:"vacancy_text"`
}