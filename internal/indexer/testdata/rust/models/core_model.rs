/// BaseEntity provides shared identity fields.
pub struct BaseEntity {
    pub id: u64,
    pub created_at: String,
}

/// CoreModel is a high-usage domain model referenced by many files.
/// Used as a fixture to verify coupling score distribution for popular symbols.
pub struct CoreModel {
    pub base: BaseEntity,
    pub title: String,
    pub score: f64,
}

impl CoreModel {
    /// Creates a new CoreModel with the given title and score.
    pub fn new(title: &str, score: f64) -> CoreModel {
        CoreModel {
            base: BaseEntity { id: 0, created_at: String::new() },
            title: title.to_string(),
            score,
        }
    }

    /// Returns true when the CoreModel fields are valid.
    pub fn is_valid(&self) -> bool {
        !self.title.is_empty() && self.score >= 0.0
    }

    /// Returns a short description of the model.
    pub fn describe(&self) -> &str {
        &self.title
    }
}
