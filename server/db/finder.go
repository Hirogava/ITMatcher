package db

import (
	"database/sql"
	"gaspr/models"
)

func (manager *Manager) CreateResumeForUser(user_id int) (int, error) {
	query := "INSERT INTO user_resumes (user_id) VALUES ($1) RETURNING id"
	var resumeId int
	err := manager.Conn.QueryRow(query, user_id).Scan(&resumeId)
	if err != nil {
		return 0, err
	}
	return resumeId, nil
}

func (manager *Manager) GetSuperDuperSecretAnonymousBitcoinWalletUnderUSAProtectionSkillAssPercentMatch(superdupersecretResumeId int, superdupersecretsupermanVacancyId int) (int, string, error) {
	query := "SELECT USA.id, v.name, USA.percent_match FROM user_skill_analysis USA INNER JOIN middle_vacancies v ON USA.vacancy_id = v.id WHERE USA.resume_id = $1 AND USA.vacancy_id = $2;"
	var USAID int
	var PercentMatch int
	var VacancyName string
	err := manager.Conn.QueryRow(query, superdupersecretResumeId, superdupersecretsupermanVacancyId).Scan(&USAID, &VacancyName, &PercentMatch)
	if err != nil {
		return 0, "", err
	}
	return PercentMatch, VacancyName, nil
}

func (manager *Manager) GetResumeUserVacancies(resumeId int) ([]int, error) {
	query := "SELECT vacancy_1_id, vacancy_2_id, vacancy_3_id FROM user_resumes WHERE id = $1"

	var vacancies [3]sql.NullInt32
	err := manager.Conn.QueryRow(query, resumeId).Scan(&vacancies[0], &vacancies[1], &vacancies[2])
	if err != nil {
		return nil, err
	}

	var result []int
	for _, vacancy := range vacancies {
		if vacancy.Valid {
			result = append(result, int(vacancy.Int32))
		}
	}

	return result, nil
}

func (manager *Manager) GetSkillsFromUserResume(resumeId int) ([]string, []string, error) {
	var hardSkills []string
	var softSkills []string
	query := `SELECT hs.hard_skill FROM user_resume_hard urh INNER JOIN hard_skills hs ON urh.hard_skill_id = hs.id WHERE urh.resume_id = $1;`

	rows, err := manager.Conn.Query(query, resumeId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hardSkill models.ResumeHardSkill
		err := rows.Scan(&hardSkill.SkillName)
		if err != nil {
			return nil, nil, err
		}
		hardSkills = append(hardSkills, hardSkill.SkillName)
	}

	query = `SELECT ss.soft_skill FROM user_resume_soft urs INNER JOIN soft_skills ss ON urs.soft_skill_id = ss.id WHERE urs.resume_id = $1;`

	rows, err = manager.Conn.Query(query, resumeId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var softSkill models.ResumeSoftSkill
		err := rows.Scan(&softSkill.SkillName)
		if err != nil {
			return nil, nil, err
		}
		softSkills = append(softSkills, softSkill.SkillName)
	}

	return hardSkills, softSkills, nil
}

func (manager *Manager) GetUserResumes(userId int) ([]models.UserResumeInfo, error) {
	query := `
		SELECT
			ur.id AS resume_id,
			v.name AS vacancy_name,
			usa.percent_match
		FROM
			user_resumes ur
		LEFT JOIN
			user_skill_analysis usa ON ur.id = usa.resume_id
		LEFT JOIN
			middle_vacancies v ON usa.vacancy_id = v.id
		WHERE
			ur.user_id = $1
		ORDER BY
			ur.id, usa.percent_match DESC
	`

	rows, err := manager.Conn.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resumeMap := make(map[int][]struct {
		Name    string
		Percent int
	})

	idMap := make(map[int]int)
	for rows.Next() {
		var resumeId int
		var vacancyName sql.NullString
		var percent sql.NullInt32

		err := rows.Scan(&resumeId, &vacancyName, &percent)
		if err != nil {
			return nil, err
		}
		idMap[resumeId]++

		if idMap[resumeId] <= 3 {
			if _, exists := resumeMap[resumeId]; !exists {
				resumeMap[resumeId] = []struct {
					Name    string
					Percent int
				}{}
			}

			if vacancyName.Valid && percent.Valid {
				resumeMap[resumeId] = append(resumeMap[resumeId], struct {
					Name    string
					Percent int
				}{
					Name:    vacancyName.String,
					Percent: int(percent.Int32),
				})
			}
		}

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var result []models.UserResumeInfo

	for resumeId, vacancies := range resumeMap {
		result = append(result, models.UserResumeInfo{
			ResumeId:  resumeId,
			Vacancies: vacancies,
		})
	}

	return result, nil
}

func (manager *Manager) UpdateUserResumesWithTopVacancies(Id int, topVacs []models.VacancyMatchResult) error {
	query := `
		UPDATE user_resumes
		SET vacancy_1_id = $1, vacancy_2_id = $2, vacancy_3_id = $3
		WHERE id = $4
	`

	var vac1, vac2, vac3 *int
	if len(topVacs) > 0 {
		vac1 = &topVacs[0].VacancyId
	}
	if len(topVacs) > 1 {
		vac2 = &topVacs[1].VacancyId
	}
	if len(topVacs) > 2 {
		vac3 = &topVacs[2].VacancyId
	}

	_, err := manager.Conn.Exec(query, vac1, vac2, vac3, Id)
	return err
}