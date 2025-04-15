package db

import (
	"database/sql"
	"fmt"
	"gaspr/models"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	Conn *sql.DB
	WG   *sync.WaitGroup
	MU   *sync.RWMutex
}

func NewDBManager(driverName string, sourceName string) *Manager {
	db, err := sql.Open(driverName, sourceName)
	if err != nil {
		panic(fmt.Sprintf("Не удалось подключиться к базе данных: %v", err))
	}

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("База данных не отвечает: %v", err))
	}

	return &Manager{
		Conn: db,
		WG:   &sync.WaitGroup{},
		MU:   &sync.RWMutex{},
	}
}

func (manager *Manager) Close() {
	if manager.Conn != nil {
		manager.Conn.Close()
		manager.Conn = nil
	}
}

/*
Users
*/
func (manager *Manager) Register(table, email, password, username string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int
	if table == "hr" {
		query := fmt.Sprintf(`INSERT INTO %s (email, hash_password, username) VALUES ($1, $2, $3) RETURNING id`, table)
		err = manager.Conn.QueryRow(query, email, hashedPassword, username).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else if table == "users" {
		query := fmt.Sprintf(`INSERT INTO %s (email, hash_password) VALUES ($1, $2) RETURNING id`, table)
		err = manager.Conn.QueryRow(query, email, hashedPassword).Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (manager *Manager) Authenticate(table, email, password string) (int, string, error) {
	var hash, username string
	var id int
	var err error

	if table == "hr" {
		err = manager.Conn.QueryRow(fmt.Sprintf(`SELECT hash_password, username, id FROM %s WHERE email=$1`, table), email).Scan(&hash, &username, &id)
	} else if table == "users" {
		err = manager.Conn.QueryRow(fmt.Sprintf(`SELECT hash_password, id FROM %s WHERE email=$1`, table), email).Scan(&hash, &id)
	}
	if err != nil {
		return 0, "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return 0, "", err
	}

	return id, username, nil
}

/*
Skills
*/
func (manager *Manager) getSkillIdByName(table, skillType, skillName string) (int, error) {
	query := fmt.Sprintf(`SELECT id FROM %s WHERE %s = $1`, table, skillType)

	var skillId int

	err := manager.Conn.QueryRow(query, skillName).Scan(&skillId)
	if err != nil {
		return 0, err
	}

	return skillId, nil
}

func (manager *Manager) GetHardSkillByName(skillName string) (int, error) {
	return manager.getSkillIdByName("hard_skills", "hard_skill", skillName)
}
func (manager *Manager) GetSoftSkillByName(skillName string) (int, error) {
	return manager.getSkillIdByName("soft_skills", "soft_skill", skillName)
}

func (manager *Manager) createSkill(tableName, skillType, skillName string) (int, error) {
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES ($1) RETURNING id`, tableName, skillType)
	var skillId int
	err := manager.Conn.QueryRow(query, skillName).Scan(&skillId)

	if err != nil {
		return 0, err
	}
	return skillId, nil
}

func (manager *Manager) CreateHardSkill(skillName string) (int, error) {
	return manager.createSkill("hard_skills", "hard_skill", skillName)
}
func (manager *Manager) CreateSoftSkill(skillName string) (int, error) {
	return manager.createSkill("soft_skills", "soft_skill", skillName)
}

/*
Resume
*/
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

func (manager *Manager) SaveAnalyzedDataForHr(resumeId int, vacancyId int, analyzedSkills models.FinalSkills) error {
	var id int
	query := "INSERT INTO hr_skill_analysis (resume_id, vacancy_id, percent_match) VALUES ($1, $2, $3) RETURNING id"
	err := manager.Conn.QueryRow(query, resumeId, vacancyId, analyzedSkills.Percent).Scan(&id)
	if err != nil {
		return err
	}

	err = manager.insertAnalyzedSkills("hr_analysis_hard_skills", id, analyzedSkills, true, "hard")
	if err != nil {
		return err
	}
	err = manager.insertAnalyzedSkills("hr_analysis_soft_skills", id, analyzedSkills, true, "soft")
	if err != nil {
		return err
	}
	err = manager.insertAnalyzedSkills("hr_analysis_hard_skills", id, analyzedSkills, false, "hard")
	if err != nil {
		return err
	}
	err = manager.insertAnalyzedSkills("hr_analysis_soft_skills", id, analyzedSkills, false, "soft")
	if err != nil {
		return err
	}

	return nil
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

/*
HR
*/

type HR struct {
	ID       int
	Username string
	Email    string
}

func (manager *Manager) GetHrInfoById(hr_id int) (HR, error) {
	var hr HR
	err := manager.Conn.QueryRow("SELECT id, username, email FROM hr WHERE id = $1", hr_id).Scan(&hr.ID, &hr.Username, &hr.Email)
	if err != nil {
		return HR{}, err
	}
	return hr, nil
}

func (manager *Manager) GetHRIdByUsername(username string) (int, error) {
	var hrId int
	err := manager.Conn.QueryRow("SELECT id FROM hr WHERE username = $1", username).Scan(&hrId)
	if err != nil {
		return 0, err
	}
	return hrId, nil
}

/*
Vacancy
*/

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

		hard, soft, err := manager.GetVacancySkills(vacancy.Id)
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

func (manager *Manager) GetAllVacancies() ([]models.Vacancy, error) {
	var vacancies []models.Vacancy

	query := `SELECT id, name FROM vacancies`
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

	vacancy.HardSkills, vacancy.SoftSkills, err = manager.GetVacancySkills(vacancy.Id)
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

func (manager *Manager) GetVacancySkills(vacancyId int) ([]models.VacancyHardSkill, []models.VacancySoftSkill, error) {
	var hardSkills []models.VacancyHardSkill
	var softSkills []models.VacancySoftSkill

	query := `SELECT hs.id, hs.hard_skill
	FROM vacantion_hard_skills vhs
	JOIN hard_skills hs ON vhs.hard_skill_id = hs.id
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

	query = `SELECT ss.id, ss.soft_skill
	FROM vacantion_soft_skills vss
	JOIN soft_skills ss ON vss.soft_skill_id = ss.id
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

/*
Finder
*/
func (manager *Manager) CreateFinder(portfolio bool, hrId int) (int, error) {
	var finderId int
	query := "INSERT INTO finders (portfolio, hr_id) VALUES ($1, $2) RETURNING id"
	err := manager.Conn.QueryRow(query, portfolio, hrId).Scan(&finderId)
	if err != nil {
		return 0, err
	}
	return finderId, nil
}

/*
User
*/
func (manager *Manager) CreateResumeForUser(user_id int) (int, error) {
	query := "INSERT INTO user_resumes (user_id) VALUES ($1) RETURNING id"
	var resumeId int
	err := manager.Conn.QueryRow(query, user_id).Scan(&resumeId)
	if err != nil {
		return 0, err
	}
	return resumeId, nil
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

func (manager *Manager) UpdateUser(role string, email string, username string, userId int) error {
	if role == "hr" {
		query := `
			UPDATE hr
			SET email = $1, username = $2
			WHERE id = $3`
		_, err := manager.Conn.Exec(query, email, username, userId)
		if err != nil {
			return err
		}
	} else if role == "users" {
		query := `
			UPDATE users
			SET email = $1
			WHERE id = $2`
		_, err := manager.Conn.Exec(query, email, userId)
		if err != nil {
			return err
		}
	}
	return nil
}
