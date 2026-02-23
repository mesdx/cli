package models

// BaseEntity provides shared identity fields for all core models.
type BaseEntity struct {
	ID        int
	CreatedAt string
}

// CoreModel is a high-usage domain model referenced by many files.
// Used as a fixture to verify coupling score distribution for popular symbols.
type CoreModel struct {
	BaseEntity
	Title string
	Score float64
}

// NewCoreModel constructs a CoreModel with defaults.
func NewCoreModel(title string, score float64) *CoreModel {
	return &CoreModel{
		BaseEntity: BaseEntity{ID: 0},
		Title:      title,
		Score:      score,
	}
}

// IsValid returns true when the CoreModel fields are non-empty.
func (m *CoreModel) IsValid() bool {
	return m.Title != "" && m.Score >= 0
}

// Describe returns a short description of the model.
func (m *CoreModel) Describe() string {
	return m.Title
}
