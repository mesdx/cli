/// type_usage_edges.rs — fixture testing that CoreModel is found as a ref
/// when it appears inside generic/container type expressions.
///
/// Every function uses CoreModel exclusively in type positions so the fixture
/// exercises Rust's type-identifier capture (`(type_identifier) @ref.type`).

use super::core_model::CoreModel;

/// Receives a Vec of CoreModel values.
pub fn takes_vec(models: Vec<CoreModel>) -> usize {
    models.len()
}

/// Receives a Vec of CoreModel references.
pub fn takes_vec_ref(models: &Vec<CoreModel>) -> usize {
    models.len()
}

/// Returns an owned Vec of CoreModel values.
pub fn returns_vec(n: usize) -> Vec<CoreModel> {
    (0..n).map(|i| CoreModel::new(&format!("item{}", i), i as f64)).collect()
}

/// Returns an Option wrapping a CoreModel.
pub fn returns_option(title: &str) -> Option<CoreModel> {
    if title.is_empty() {
        None
    } else {
        Some(CoreModel::new(title, 1.0))
    }
}

/// Returns a Result whose Ok variant is a CoreModel.
pub fn returns_result(title: &str) -> Result<CoreModel, String> {
    if title.is_empty() {
        Err("title empty".to_string())
    } else {
        Ok(CoreModel::new(title, 1.0))
    }
}

/// Box<CoreModel> as a return type.
pub fn returns_boxed() -> Box<CoreModel> {
    Box::new(CoreModel::new("boxed", 0.5))
}

/// Option<Box<CoreModel>> as a return type (doubly nested generic).
pub fn returns_option_boxed() -> Option<Box<CoreModel>> {
    Some(Box::new(CoreModel::new("opt_boxed", 0.7)))
}

/// Local variable declared with type CoreModel.
pub fn local_var(title: &str) -> bool {
    let m: CoreModel = CoreModel::new(title, 0.5);
    m.is_valid()
}

/// Tuple type containing CoreModel.
pub fn returns_tuple(title: &str) -> (CoreModel, bool) {
    let m = CoreModel::new(title, 1.0);
    let valid = m.is_valid();
    (m, valid)
}

/// Accepts a closure that receives and returns CoreModel.
pub fn apply<F>(model: CoreModel, f: F) -> CoreModel
where
    F: Fn(CoreModel) -> CoreModel,
{
    f(model)
}
