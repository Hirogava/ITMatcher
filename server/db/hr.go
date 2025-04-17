package db

import "gaspr/models"

func (manager *Manager) GetHrInfoById(hr_id int) (models.HR, error) {
	var hr models.HR
	err := manager.Conn.QueryRow("SELECT id, username, email FROM hr WHERE id = $1", hr_id).Scan(&hr.ID, &hr.Username, &hr.Email)
	if err != nil {
		return models.HR{}, err
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

func (manager *Manager) CreateFinder(portfolio bool, hrId int) (int, error) {
	var finderId int
	query := "INSERT INTO finders (portfolio, hr_id) VALUES ($1, $2) RETURNING id"
	err := manager.Conn.QueryRow(query, portfolio, hrId).Scan(&finderId)
	if err != nil {
		return 0, err
	}
	return finderId, nil
}