"""Tricky Python fixture for deep attribute analysis."""

import os
import json
from typing import Dict, List, Optional, Any
from collections import defaultdict
from dataclasses import dataclass, field

# --- Shadowing: local shadows global ---

MAX_RETRIES = 5  # module-level constant


def shadow_example():
    MAX_RETRIES = 10  # shadows module-level
    print(MAX_RETRIES)


# --- Multiple inheritance ---


class Serializable:
    def serialize(self) -> str:
        return json.dumps(self.__dict__)


class Printable:
    def display(self) -> None:
        print(str(self))


class Document(Serializable, Printable):
    def __init__(self, title: str, content: str):
        self.title = title
        self.content = content

    def __str__(self) -> str:
        return f"Document({self.title})"


# --- Decorators and properties ---


class Config:
    def __init__(self, data: Dict[str, Any]):
        self._data = data
        self._cache: Dict[str, Any] = {}

    @property
    def host(self) -> str:
        return self._data.get("host", "localhost")

    @host.setter
    def host(self, value: str) -> None:
        self._data["host"] = value

    @staticmethod
    def from_env() -> "Config":
        return Config({"host": os.getenv("HOST", "localhost")})

    @classmethod
    def default(cls) -> "Config":
        return cls({"host": "localhost", "port": 8080})

    # --- Nested function ---
    def process(self, items: List[str]) -> List[str]:
        def transform(item: str) -> str:
            return item.strip().lower()

        return [transform(i) for i in items]


# --- Dataclass ---


@dataclass
class Point:
    x: float
    y: float
    label: str = ""
    metadata: Dict[str, Any] = field(default_factory=dict)

    def distance_to(self, other: "Point") -> float:
        return ((self.x - other.x) ** 2 + (self.y - other.y) ** 2) ** 0.5


# --- Builtins usage (should NOT be unresolved external) ---


def use_builtins():
    # builtin functions
    x = len([1, 2, 3])
    y = range(10)
    z = list(y)
    s = str(x)
    i = int("42")
    f = float("3.14")
    b = bool(1)
    d = dict(a=1, b=2)
    t = tuple([1, 2])
    st = set([1, 2, 3])
    _ = isinstance(x, int)
    _ = issubclass(int, object)
    _ = hasattr([], "append")
    _ = getattr([], "append")
    _ = callable(print)
    _ = type(x)
    _ = repr(x)
    _ = hash(42)
    _ = abs(-1)
    _ = min(1, 2)
    _ = max(1, 2)
    _ = sorted([3, 1, 2])
    _ = reversed([1, 2, 3])
    _ = enumerate(z)
    _ = zip([1], [2])
    _ = map(str, [1, 2])
    _ = filter(None, [0, 1])
    _ = any([True, False])
    _ = all([True, True])
    print(s, i, f, b, d, t, st)
    raise ValueError("test")


# --- External stdlib references ---


def use_external_stdlib():
    # os module
    cwd = os.getcwd()
    home = os.environ.get("HOME", "/tmp")

    # json module
    data = json.dumps({"key": "value"})
    parsed = json.loads(data)

    # collections
    counter = defaultdict(int)
    counter["a"] += 1

    return cwd, home, parsed, counter


# --- Name collision: function and class with same prefix ---

Task = None  # module-level variable


class Task:
    """Class with same name as module-level variable."""

    def __init__(self, name: str):
        self.name = name

    def run(self) -> None:
        pass


# --- Dunder methods ---


class Container:
    def __init__(self):
        self._items: List[Any] = []

    def __len__(self) -> int:
        return len(self._items)

    def __getitem__(self, index: int) -> Any:
        return self._items[index]

    def __setitem__(self, index: int, value: Any) -> None:
        self._items[index] = value

    def __contains__(self, item: Any) -> bool:
        return item in self._items

    def __iter__(self):
        return iter(self._items)


# --- Import alias usage ---

dd = defaultdict  # alias for defaultdict
