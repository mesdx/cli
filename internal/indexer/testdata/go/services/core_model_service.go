package services

import "models"

// CoreModelRepository stores and retrieves CoreModel instances.
// Demonstrates instantiation (high coupling) and method call (medium coupling).
type CoreModelRepository struct {
	store []models.CoreModel
}

// NewCoreModelRepository creates a populated repository with two CoreModel entries.
// Coupling: struct literal instantiation → high-coupling usage of CoreModel.
func NewCoreModelRepository() *CoreModelRepository {
	first := &models.CoreModel{Title: "alpha", Score: 1.0}
	second := models.CoreModel{Title: "beta", Score: 0.5}
	return &CoreModelRepository{
		store: []models.CoreModel{*first, second},
	}
}

// Add stores a CoreModel by instantiating from raw fields.
// Coupling: struct literal + method call → high coupling.
func (r *CoreModelRepository) Add(title string, score float64) {
	m := &models.CoreModel{Title: title, Score: score}
	if m.IsValid() {
		r.store = append(r.store, *m)
	}
}

// Find returns a CoreModel matching the title.
func (r *CoreModelRepository) Find(title string) *models.CoreModel {
	for i := range r.store {
		if r.store[i].Title == title {
			return &r.store[i]
		}
	}
	return nil
}

// Describe calls Describe on each CoreModel and returns the first non-empty result.
// Coupling: method call → medium coupling.
func (r *CoreModelRepository) Describe() string {
	for i := range r.store {
		if d := r.store[i].Describe(); d != "" {
			return d
		}
	}
	return ""
}

// StoreCoreModel persists a CoreModel using a standalone function.
func StoreCoreModel(m models.CoreModel) bool {
	return m.IsValid()
}

// LoadCoreModel fabricates a CoreModel by value — another instantiation.
func LoadCoreModel(title string) models.CoreModel {
	return models.CoreModel{Title: title, Score: 0.0}
}
