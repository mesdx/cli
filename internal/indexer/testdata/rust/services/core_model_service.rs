use crate::models::core_model::CoreModel;

/// CoreModelRepository stores and retrieves CoreModel instances.
/// Demonstrates: use-import (0.40), struct-field composition (0.25),
/// CoreModel::new instantiation (0.95 via lexical), method calls (0.65).
pub struct CoreModelRepository {
    store: Vec<CoreModel>,
}

impl CoreModelRepository {
    /// Creates a new repository seeded with an initial CoreModel.
    /// Coupling: CoreModel::new → instantiation (lexical escalation → 0.95).
    pub fn new() -> CoreModelRepository {
        let initial = CoreModel::new("initial", 0.5);
        CoreModelRepository { store: vec![initial] }
    }

    /// add stores a CoreModel in the repository.
    pub fn add(&mut self, model: CoreModel) {
        if model.is_valid() {
            self.store.push(model);
        }
    }

    /// find_by_title returns the first CoreModel matching the title.
    pub fn find_by_title(&self, title: &str) -> Option<&CoreModel> {
        self.store.iter().find(|m| m.title == title)
    }

    /// describe calls describe() on each CoreModel.
    pub fn describe(&self) -> Vec<&str> {
        self.store.iter().map(|m| m.describe()).collect()
    }
}

/// store_core_model persists a CoreModel using a standalone function.
pub fn store_core_model(model: CoreModel) -> bool {
    model.is_valid()
}

/// build_core_model uses the CoreModel constructor — another instantiation pattern.
pub fn build_core_model(title: &str, score: f64) -> CoreModel {
    CoreModel::new(title, score)
}
