"""
type_hint_edges.py — fixture for type-annotation coverage tests.

Exercises the three Python annotation patterns where the symbol name is buried
inside a string literal and therefore invisible to the plain identifier
capture:

  Pattern 1 – direct quoted return type:
      def foo() -> "DataModel": ...

  Pattern 2 – quoted element inside generic (subscript):
      def bar() -> List["DataModel"]: ...
      def baz() -> tuple["DataModel", bool]: ...

  Pattern 3 – fully-quoted complex type expression:
      def qux() -> "tuple[DataModel, bool]": ...

All three patterns are present as both return-type annotations and
parameter-type annotations so the extractor captures are exercised fully.

Negative controls: ordinary string variables and docstrings that mention
"DataModel" must NOT produce annotation refs.
"""

from __future__ import annotations

from typing import List, Optional, Union, Tuple


class DataModel:
    """The symbol under test.  All type_hint_edge functions reference it."""

    def __init__(self, value: int) -> None:
        self.value = value

    def is_valid(self) -> bool:
        return self.value >= 0


# ---------------------------------------------------------------------------
# Pattern 1 – direct quoted return/parameter type
# ---------------------------------------------------------------------------

def get_model_direct_quoted() -> "DataModel":
    """Return type is a plain quoted identifier."""
    return DataModel(1)


def process_model_param_quoted(model: "DataModel") -> bool:
    """Parameter type is a plain quoted identifier."""
    return model.is_valid()


def get_optional_quoted() -> "Optional[DataModel]":
    """Return type: quoted expression wrapping Optional."""
    return DataModel(2)


# ---------------------------------------------------------------------------
# Pattern 2 – quoted element inside a generic container (subscript)
# ---------------------------------------------------------------------------

def get_list_of_models_generic() -> List["DataModel"]:
    """Generic return: List with a quoted element."""
    return [DataModel(3)]


def get_tuple_with_quoted() -> tuple["DataModel", bool]:
    """Generic tuple return with a quoted element (Python 3.9+ style)."""
    m = DataModel(4)
    return (m, m.is_valid())


def get_optional_generic() -> Optional["DataModel"]:
    """Optional with a quoted element."""
    return DataModel(5)


def get_union_generic() -> Union["DataModel", None]:
    """Union with a quoted element."""
    return DataModel(6)


def process_list_param(models: List["DataModel"]) -> int:
    """Parameter type: List with a quoted element."""
    return len(models)


def process_tuple_param(data: tuple["DataModel", bool]) -> "DataModel":
    """Both parameter and return type use quoted element.  Double hit."""
    return data[0]


# ---------------------------------------------------------------------------
# Pattern 3 – the entire annotation is a quoted complex expression
# ---------------------------------------------------------------------------

def get_model_fully_quoted() -> "DataModel":
    """Return type is a single-token quoted forward reference."""
    return DataModel(7)


def get_tuple_fully_quoted() -> "tuple[DataModel, bool]":
    """Entire tuple expression is quoted; parser must find DataModel inside."""
    m = DataModel(8)
    return (m, True)


def get_list_fully_quoted() -> "List[DataModel]":
    """Entire List expression is quoted; parser must find DataModel inside."""
    return [DataModel(9)]


# ---------------------------------------------------------------------------
# Pattern 4 – deeply nested generics
# ---------------------------------------------------------------------------

def get_nested_generic() -> Optional[List["DataModel"]]:
    """DataModel is two levels deep: Optional > List > "DataModel"."""
    return [DataModel(10)]


def get_deeply_nested() -> Optional["List[DataModel]"]:
    """DataModel is inside a quoted expression which is itself inside Optional."""
    return [DataModel(11)]


# ---------------------------------------------------------------------------
# Variable and return type annotations (PEP 526 / PEP 563)
# ---------------------------------------------------------------------------

latest: "DataModel" = DataModel(12)

cache: List["DataModel"] = []

# ---------------------------------------------------------------------------
# Negative controls — must NOT produce DataModel annotation refs
# ---------------------------------------------------------------------------

# Plain string that mentions the class name (not in annotation position)
description = "Please use DataModel for all data."

# Docstring (not an annotation)
"""
Here is how to use DataModel in your code.
"""


def plain_string_not_annotation() -> str:
    """Function that returns a plain string mentioning DataModel."""
    return "DataModel is the key class"


def string_argument_not_annotation(key: str = "DataModel") -> None:
    """Default argument value is a string, NOT an annotation."""
    pass


# Multi-assignment with string on right-hand side — not an annotation
model_name_str = "DataModel"
