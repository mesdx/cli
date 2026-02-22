from models.critical_user_model import CriticalUserModel


class CriticalUserModelRepository:
    """Manages CriticalUserModel persistence."""

    def __init__(self):
        self._store: list = []

    def save(self, model: CriticalUserModel) -> bool:
        self._store.append(model)
        return True

    def find_by_email(self, email: str):
        for m in self._store:
            if isinstance(m, CriticalUserModel) and m.email == email:
                return m
        return None

    def find_all(self) -> list:
        return list(self._store)


def store_critical_user_model(model: CriticalUserModel) -> bool:
    """Standalone function for persisting a CriticalUserModel."""
    return model.validate()
