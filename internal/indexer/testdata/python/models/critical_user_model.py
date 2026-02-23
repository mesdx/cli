"""Critical user model module."""


class BaseModel:
    """BaseModel provides base entity identity fields."""

    def __init__(self, id: int = 0):
        self.id = id
        self.created_at = ""


class CriticalUserModel(BaseModel):
    """CriticalUserModel is a heavily-used user representation that many files depend on."""

    def __init__(self, name: str, email: str):
        super().__init__(0)
        self.name = name
        self.email = email

    def validate(self) -> bool:
        return bool(self.name) and "@" in self.email

    def display(self) -> str:
        return f"{self.name} <{self.email}>"


def create_critical_user_model(name: str, email: str) -> CriticalUserModel:
    """Factory function for creating CriticalUserModel instances."""
    return CriticalUserModel(name, email)
