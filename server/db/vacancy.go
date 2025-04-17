package db

import (
	"fmt"
	"gaspr/models"
)

func (manager *Manager) CreateVacancy(name string, hr_id int) (int, error) {
	var vacId int
	err := manager.Conn.QueryRow("INSERT INTO vacancies (name, hr_id) VALUES ($1, $2) RETURNING id", name, hr_id).Scan(&vacId)
	if err != nil {
		return 0, err
	}
	return vacId, err
}

func (manager *Manager) GetAllHrVacancies(hr_id int) ([]models.Vacancy, error) {
	var vacancies []models.Vacancy

	query := `SELECT v.id, v.name
	FROM vacancies v
	WHERE v.hr_id = $1`
	rows, err := manager.Conn.Query(query, hr_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var vacancy models.Vacancy
		err := rows.Scan(&vacancy.Id, &vacancy.Name)
		if err != nil {
			return nil, err
		}

		hard, soft, err := manager.GetVacancySkills(vacancy.Id, "hr")
		if err != nil {
			return nil, err
		}

		vacancy.HardSkills = hard
		vacancy.SoftSkills = soft
		vacancies = append(vacancies, vacancy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vacancies, nil
}

func (manager *Manager) GetAllVacancies(role string) ([]models.Vacancy, error) {
	var vacancies []models.Vacancy
	var query string

	if role == "hr" {
		query = `SELECT id, name FROM vacancies`
	} else if role == "users" {
		query = `SELECT id, name FROM middle_vacancies`
	}
	rows, err := manager.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var vacancy models.Vacancy
		err := rows.Scan(&vacancy.Id, &vacancy.Name)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, vacancy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return vacancies, nil
}

func (manager *Manager) GetVacancyIdByName(name string, hr_id int) (int, error) {
	var vacId int
	err := manager.Conn.QueryRow("SELECT id FROM vacancies WHERE name = $1 AND hr_id = $2", name, hr_id).Scan(&vacId)
	if err != nil {
		return 0, err
	}
	return vacId, nil
}

func (manager *Manager) GetVacancyByIdForHr(id int) (models.Vacancy, error) {
	var vacancy models.Vacancy

	query := "SELECT id, name FROM vacancies WHERE id = $1"
	err := manager.Conn.QueryRow(query, id).Scan(&vacancy.Id, &vacancy.Name)
	if err != nil {
		return vacancy, err
	}

	vacancy.HardSkills, vacancy.SoftSkills, err = manager.GetVacancySkills(vacancy.Id, "hr")
	if err != nil {
		return vacancy, err
	}

	return vacancy, nil
}

func (manager *Manager) createVacancySkill(tableName, skillType string, vacancyId, skillId int) error {
	query := fmt.Sprintf(`INSERT INTO %s (vacancy_id, %s) VALUES ($1, $2)`, tableName, skillType)

	_, err := manager.Conn.Exec(query, vacancyId, skillId)
	return err
}

func (manager *Manager) CreateVacancyHardSkill(vacancyId int, skillId int) error {
	return manager.createVacancySkill("vacantion_hard_skills", "hard_skill_id", vacancyId, skillId)
}
func (manager *Manager) CreateVacancySoftSkill(vacancyId int, skillId int) error {
	return manager.createVacancySkill("vacantion_soft_skills", "soft_skill_id", vacancyId, skillId)
}

func (manager *Manager) GetVacancySkills(vacancyId int, role string) ([]models.VacancyHardSkill, []models.VacancySoftSkill, error) {
	var hardSkills []models.VacancyHardSkill
	var softSkills []models.VacancySoftSkill
	query := `SELECT hs.id, hs.hard_skill`

	if role == "hr" {
		query += ` FROM vacantion_hard_skills vhs `
	} else if role == "users" {
		query += ` FROM middle_hard_skills vhs `
	}
	query += `JOIN hard_skills hs ON vhs.hard_skill_id = hs.id
		WHERE vhs.vacancy_id = $1`

	rows, err := manager.Conn.Query(query, vacancyId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hardSkill models.VacancyHardSkill
		err := rows.Scan(&hardSkill.Id, &hardSkill.SkillName)
		if err != nil {
			return nil, nil, err
		}
		hardSkills = append(hardSkills, hardSkill)
	}

	query = `SELECT ss.id, ss.soft_skill`
	if role == "hr" {
		query += ` FROM vacantion_soft_skills vss `
	} else if role == "users" {
		query += ` FROM middle_soft_skills vss `
	}
	query += `JOIN soft_skills ss ON vss.soft_skill_id = ss.id
	WHERE vss.vacancy_id = $1`

	rows, err = manager.Conn.Query(query, vacancyId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var softSkill models.VacancySoftSkill
		err := rows.Scan(&softSkill.Id, &softSkill.SkillName)
		if err != nil {
			return nil, nil, err
		}
		softSkills = append(softSkills, softSkill)
	}

	return hardSkills, softSkills, nil
}