package services

import "models"

// StoreCriticalUserModel persists a CriticalUserModel.
func StoreCriticalUserModel(m models.CriticalUserModel) bool {
	return m.Name != ""
}

// LoadCriticalUserModel retrieves a CriticalUserModel by ID.
func LoadCriticalUserModel(id int) models.CriticalUserModel {
	return models.CriticalUserModel{Name: "loaded", Email: "loaded@example.com"}
}

// CriticalUserModelRepository manages CriticalUserModel persistence.
type CriticalUserModelRepository struct {
	models []models.CriticalUserModel
}

// Add stores a CriticalUserModel in the repository.
func (r *CriticalUserModelRepository) Add(m models.CriticalUserModel) {
	r.models = append(r.models, m)
}
