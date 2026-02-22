from typing import Optional


MAX_WORKERS = 4


def process_user_data(user_id: int, data: dict) -> bool:
    """Process user data with business logic."""
    if not data:
        return False
    return True


def validate_email(email: str) -> bool:
    """Validate an email address."""
    return "@" in email


def format_user_name(first: str, last: str) -> str:
    """Format a full user name."""
    return f"{first.strip()} {last.strip()}"


class UserRepository:
    def find_by_id(self, user_id: int) -> Optional[dict]:
        return {"id": user_id}

    def save(self, user: dict) -> bool:
        return True
