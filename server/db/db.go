package db

import (
	"database/sql"
	"fmt"
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

	query := fmt.Sprintf(`INSERT INTO %s (email, hash_password, username) VALUES ($1, $2, $3) RETURNING id`, table)
	var id int
	err = manager.Conn.QueryRow(query, email, hashedPassword, username).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (manager *Manager) Authenticate(table, email, password string) (int, string, error) {
	var hash, username string
	var id int
	err := manager.Conn.QueryRow(fmt.Sprintf(`SELECT hash_password, username, id FROM %s WHERE email=$1`, table), email).Scan(&hash, &username, &id)
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
	query := `SELECT r.id, r.finder_id, r.first_name, r.last_name, r.surname, r.email, r.phone_number, r.vacancy_id
	FROM resumes r
	JOIN finders f ON r.finder_id = f.id
	WHERE f.hr_id = $1`
	rows, err := manager.Conn.Query(query, hr_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resume Resume
		err := rows.Scan(&resume.Id, &resume.FinderId, &resume.FirstName, &resume.LastName, &resume.Surname, &resume.Email, &resume.PhoneNumber, &resume.VacancyId)
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
func (manager *Manager) CreateVacancy(name string) (int, error) {
	var vacId int
	err := manager.Conn.QueryRow("INSERT INTO vacancies (name) VALUES ($1) RETURNING id", name).Scan(&vacId)
	if err != nil {
		return 0, err
	}
	return vacId, err
}

func (manager *Manager) GetVacancyIdByName(name string) (int, error) {
	var vacId int
	err := manager.Conn.QueryRow("SELECT id FROM vacancies WHERE name = $1", name).Scan(&vacId)
	if err != nil {
		return 0, err
	}
	return vacId, nil
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
