package models

// BaseModel provides base entity identity fields.
type BaseModel struct {
	ID        int
	CreatedAt string
}

// CriticalUserModel is a heavily-used user representation that many files depend on.
type CriticalUserModel struct {
	BaseModel
	Name  string
	Email string
}

// NewCriticalUserModel creates a new CriticalUserModel with defaults.
func NewCriticalUserModel(name, email string) *CriticalUserModel {
	return &CriticalUserModel{
		BaseModel: BaseModel{ID: 0},
		Name:      name,
		Email:     email,
	}
}

// IsValid checks that the CriticalUserModel fields are non-empty.
func (m *CriticalUserModel) IsValid() bool {
	return m.Name != "" && m.Email != ""
}
