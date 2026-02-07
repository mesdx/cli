from typing import List


# AppProcessor handles data processing.
class AppProcessor:
    def __init__(self, name: str, workers: int = 1):
        self.name = name
        self.workers = workers

    def run(self):
        for i in range(self.workers):
            print(f"Worker {i} started for {self.name}")

    def stop(self):
        print(f"Stopping {self.name}")


def format_data(items: List[str]) -> str:
    return ", ".join(items)
