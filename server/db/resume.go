package db

import (
	"fmt"
	"gaspr/models"
)

type Resume struct {
	Id          int
	FinderId    int
	FirstName   string
	LastName    string
	Surname     string
	Email       string
	PhoneNumber string
	VacancyId   int
	Percent     int
}

func (manager *Manager) GetResumeByIdForHr(resumeId int) (*Resume, error) {
	var resume Resume
	query := "SELECT finder_id, first_name, last_name, surname, email, phone_number, vacancy_id FROM resumes WHERE id = $1"
	err := manager.Conn.QueryRow(query, resumeId).Scan(&resume.Id, &resume.FinderId, &resume.FirstName, &resume.LastName, &resume.Surname, &resume.Email, &resume.PhoneNumber, &resume.VacancyId)
	if err != nil {
		return nil, err
	}
	return &resume, nil
}

func (manager *Manager) GetAllResumesForHr(hr_id int) ([]Resume, error) {
	var resumes []Resume
	query := `
		SELECT r.id, r.finder_id, r.first_name, r.last_name, r.surname, r.email, r.phone_number, r.vacancy_id,
		       COALESCE(hsa.percent_match, 0) AS percent_match
		FROM resumes r
		JOIN finders f ON r.finder_id = f.id
		LEFT JOIN hr_skill_analysis hsa 
		    ON hsa.resume_id = r.id AND hsa.vacancy_id = r.vacancy_id
		WHERE f.hr_id = $1
	`
	rows, err := manager.Conn.Query(query, hr_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resume Resume
		err := rows.Scan(
			&resume.Id,
			&resume.FinderId,
			&resume.FirstName,
			&resume.LastName,
			&resume.Surname,
			&resume.Email,
			&resume.PhoneNumber,
			&resume.VacancyId,
			&resume.Percent,
		)
		if err != nil {
			return nil, err
		}
		resumes = append(resumes, resume)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return resumes, nil
}

func (manager *Manager) SaveAnalyzedData(role string, resumeId int, vacancyId int, analyzedSkills models.FinalSkills) error {
	var id int
	if role == "users" {
		role = "user"
	}
	query := fmt.Sprintf("INSERT INTO %s_skill_analysis (resume_id, vacancy_id, percent_match) VALUES ($1, $2, $3) RETURNING id", role)
	err := manager.Conn.QueryRow(query, resumeId, vacancyId, analyzedSkills.Percent).Scan(&id)
	if err != nil {
		return err
	}

	err = manager.insertAnalyzedSkills(fmt.Sprintf("%s_analysis_hard_skills", role), id, analyzedSkills, true, "hard")
	if err != nil {
		return err
	}
	err = manager.insertAnalyzedSkills(fmt.Sprintf("%s_analysis_soft_skills", role), id, analyzedSkills, true, "soft")
	if err != nil {
		return err
	}

	err = manager.insertAnalyzedSkills(fmt.Sprintf("%s_analysis_hard_skills", role), id, analyzedSkills, false, "hard")
	if err != nil {
		return err
	}

	err = manager.insertAnalyzedSkills(fmt.Sprintf("%s_analysis_soft_skills", role), id, analyzedSkills, false, "soft")
	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) GetAnalizedUserData(resumeId int, vacancyId int) (models.AnalyzedSkills, error) {
	query := `
	(
	  SELECT hs.hard_skill AS skill, ahs.matched
	  FROM user_analysis_hard_skills ahs
	  JOIN hard_skills hs ON ahs.hard_skill_id = hs.id
	  JOIN user_skill_analysis hsa ON ahs.analysis_id = hsa.id
	  WHERE hsa.resume_id = $1 AND hsa.vacancy_id = $2
	)
	UNION ALL
	(
	  SELECT ss.soft_skill AS skill, zhopa.matched
	  FROM user_analysis_soft_skills zhopa
	  JOIN soft_skills ss ON zhopa.soft_skill_id = ss.id
	  JOIN user_skill_analysis hsa ON zhopa.analysis_id = hsa.id
	  WHERE hsa.resume_id = $1 AND hsa.vacancy_id = $2
	)
	`

	rows, err := manager.Conn.Query(query, resumeId, vacancyId)
	if err != nil {
		return models.AnalyzedSkills{}, err
	}
	defer rows.Close()

	var result models.AnalyzedSkills

	for rows.Next() {
		var skill string
		var matched bool

		if err := rows.Scan(&skill, &matched); err != nil {
			return models.AnalyzedSkills{}, err
		}

		if matched {
			result.Coincidence = append(result.Coincidence, skill)
		} else {
			result.Mismatch = append(result.Mismatch, skill)
		}
	}

	if err = rows.Err(); err != nil {
		return models.AnalyzedSkills{}, err
	}

	return result, nil
}

func (manager *Manager) GetAnalizedData(finderId int, vacancyId int) (models.AnalyzedSkills, error) {
	query := `
	(
	  SELECT hs.hard_skill AS skill, ahs.matched
	  FROM hr_analysis_hard_skills ahs
	  JOIN hard_skills hs ON ahs.hard_skill_id = hs.id
	  JOIN hr_skill_analysis hsa ON ahs.analysis_id = hsa.id
	  WHERE hsa.resume_id = $1 AND hsa.vacancy_id = $2
	)
	UNION ALL
	(
	  SELECT ss.soft_skill AS skill, ass.matched
	  FROM hr_analysis_soft_skills ass
	  JOIN soft_skills ss ON ass.soft_skill_id = ss.id
	  JOIN hr_skill_analysis hsa ON ass.analysis_id = hsa.id
	  WHERE hsa.resume_id = $1 AND hsa.vacancy_id = $2
	)
	`
	rows, err := manager.Conn.Query(query, finderId, vacancyId)
	if err != nil {
		return models.AnalyzedSkills{}, err
	}
	defer rows.Close()

	var result models.AnalyzedSkills

	for rows.Next() {
		var skill string
		var matched bool

		if err := rows.Scan(&skill, &matched); err != nil {
			return models.AnalyzedSkills{}, err
		}

		if matched {
			result.Coincidence = append(result.Coincidence, skill)
		} else {
			result.Mismatch = append(result.Mismatch, skill)
		}
	}

	if err = rows.Err(); err != nil {
		return models.AnalyzedSkills{}, err
	}

	return result, nil
}

func (manager *Manager) insertAnalyzedSkills(table string, id int, analyzedSkills models.FinalSkills, matched bool, skillType string) error {
	var skillIds []int

	switch skillType {
	case "hard":
		if matched {
			for _, skill := range analyzedSkills.CoincidenceHard {
				skillIds = append(skillIds, skill.Id)
			}
		} else {
			for _, skill := range analyzedSkills.MismatchHard {
				skillIds = append(skillIds, skill.Id)
			}
		}
	case "soft":
		if matched {
			for _, skill := range analyzedSkills.CoincidenceSoft {
				skillIds = append(skillIds, skill.Id)
			}
		} else {
			for _, skill := range analyzedSkills.MismatchSoft {
				skillIds = append(skillIds, skill.Id)
			}
		}
	default:
		return fmt.Errorf("неизвестный тип скилла: %s", skillType)
	}

	for _, skillId := range skillIds {
		query := fmt.Sprintf("INSERT INTO %s (analysis_id, %s_skill_id, matched) VALUES ($1, $2, $3)", table, skillType)
		_, err := manager.Conn.Exec(query, id, skillId, matched)
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *Manager) CreateResumeForHr(finderId int, firstName, lastName, surName, email, phoneNumber string, vacancyId int) (int, error) {
	var resumeId int
	query := "INSERT INTO resumes (finder_id, first_name, last_name, surname, email, phone_number, vacancy_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	err := manager.Conn.QueryRow(query, finderId, firstName, lastName, surName, email, phoneNumber, vacancyId).Scan(&resumeId)
	if err != nil {
		return 0, err
	}
	return resumeId, nil
}

func (manager *Manager) createResumeSkill(tableName, skillType string, resumeId, skillId int) error {
	query := fmt.Sprintf(`INSERT INTO %s (resume_id, %s) VALUES ($1, $2)`, tableName, skillType)

	_, err := manager.Conn.Exec(query, resumeId, skillId)
	return err
}

func (manager *Manager) CreateResumeHardSkill(resumeId int, skillId int) error {
	return manager.createResumeSkill("resume_hard_skill", "hard_skill_id", resumeId, skillId)
}
func (manager *Manager) CreateResumeSoftSkill(resumeId int, skillId int) error {
	return manager.createResumeSkill("resume_soft_skill", "soft_skill_id", resumeId, skillId)
}

func (manager *Manager) CreateUserResumeHardSkill(resumeId int, skillId int) error {
	return manager.createResumeSkill("user_resume_hard", "hard_skill_id", resumeId, skillId)
}
func (manager *Manager) CreateUserResumeSoftSkill(resumeId int, skillId int) error {
	return manager.createResumeSkill("user_resume_soft", "soft_skill_id", resumeId, skillId)
}