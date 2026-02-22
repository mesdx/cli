from models.critical_user_model import CriticalUserModel, create_critical_user_model


class CriticalUserModelCache:
    """Caches CriticalUserModel instances by id."""

    def __init__(self):
        self._entries: dict = {}

    def put(self, id: int, model: CriticalUserModel) -> None:
        self._entries[id] = model

    def get(self, id: int):
        return self._entries.get(id)

    def values(self) -> list:
        return list(self._entries.values())


def transform_critical_user_model(model: CriticalUserModel) -> str:
    """Transforms a CriticalUserModel into a tagged string."""
    return f"[{model.name}]"
