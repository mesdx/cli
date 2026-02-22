from models.critical_user_model import CriticalUserModel, create_critical_user_model


class CriticalUserModelRenderer:
    """Renders a CriticalUserModel for display."""

    def __init__(self, model: CriticalUserModel):
        self.model = model

    def render(self) -> str:
        return self.model.display()


def render_critical_user_model(m: CriticalUserModel) -> str:
    """Formats a CriticalUserModel as a display string."""
    return f"{m.name} <{m.email}>"


def build_critical_user_model(name: str, email: str) -> CriticalUserModel:
    """Creates a CriticalUserModel from the given parts."""
    return create_critical_user_model(name, email)
