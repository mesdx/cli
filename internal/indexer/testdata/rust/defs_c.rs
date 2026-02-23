use crate::models::critical_user_model::CriticalUserModel;

/// CriticalUserModelCache caches CriticalUserModel instances by id.
pub struct CriticalUserModelCache {
    entries: Vec<(u64, CriticalUserModel)>,
}

impl CriticalUserModelCache {
    pub fn new() -> CriticalUserModelCache {
        CriticalUserModelCache { entries: Vec::new() }
    }

    /// Inserts a CriticalUserModel with the given id.
    pub fn insert(&mut self, id: u64, model: CriticalUserModel) {
        self.entries.push((id, model));
    }

    /// Looks up a CriticalUserModel by id.
    pub fn get(&self, id: u64) -> Option<&CriticalUserModel> {
        self.entries.iter().find(|(k, _)| *k == id).map(|(_, v)| v)
    }
}
