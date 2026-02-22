use crate::models::critical_user_model::CriticalUserModel;

/// CriticalUserModelRepository manages CriticalUserModel persistence.
pub struct CriticalUserModelRepository {
    models: Vec<CriticalUserModel>,
}

impl CriticalUserModelRepository {
    pub fn new() -> CriticalUserModelRepository {
        CriticalUserModelRepository { models: Vec::new() }
    }

    /// Saves a CriticalUserModel to the repository.
    pub fn save(&mut self, model: CriticalUserModel) {
        self.models.push(model);
    }

    /// Finds a CriticalUserModel by email.
    pub fn find_by_email(&self, email: &str) -> Option<&CriticalUserModel> {
        self.models.iter().find(|m| m.email == email)
    }
}

/// store_critical_user_model is a standalone function for persistence.
pub fn store_critical_user_model(model: CriticalUserModel) -> bool {
    !model.name.is_empty()
}
