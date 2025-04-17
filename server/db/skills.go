package db

import "fmt"

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