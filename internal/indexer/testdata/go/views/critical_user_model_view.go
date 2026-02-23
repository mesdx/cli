package views

import "models"

// CriticalUserModelRenderer renders a CriticalUserModel for display.
type CriticalUserModelRenderer struct {
	Model models.CriticalUserModel
}

// RenderCriticalUserModel formats a CriticalUserModel as a display string.
func RenderCriticalUserModel(m models.CriticalUserModel) string {
	return m.Name + " <" + m.Email + ">"
}

// CreateCriticalUserModelView creates a renderer for a CriticalUserModel.
func CreateCriticalUserModelView(name, email string) models.CriticalUserModel {
	return models.CriticalUserModel{Name: name, Email: email}
}
