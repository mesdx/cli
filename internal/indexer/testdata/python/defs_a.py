from typing import Optional

# Default port for the application.
DEFAULT_PORT = 8080

# Application name.
app_name = "myapp"


# AppConfig holds application configuration.
class AppConfig:
    def __init__(self, host: str, port: int = DEFAULT_PORT):
        self.host = host
        self.port = port

    def validate(self) -> bool:
        if not self.host:
            raise ValueError("host is required")
        if self.port <= 0:
            raise ValueError("port must be positive")
        return True

    @property
    def address(self) -> str:
        return f"{self.host}:{self.port}"


class AppFormatter:
    """Formatter for output formatting."""

    def format(self, data: str) -> str:
        return f"[{data}]"

    def reset(self):
        pass


# say_hello is a standalone function.
def say_hello(name: str) -> str:
    greeting = f"Hello, {name}!"
    print(greeting)
    return greeting
