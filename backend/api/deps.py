from typing import Generator
from backend.services.interfaces import ICodeAnalyzer
from backend.services.openai_service import OpenAIAnalyzer

# Dependency Injection
def get_analyzer() -> Generator[ICodeAnalyzer, None, None]:
    # Here we can easily swap OpenAIAnalyzer for another implementation
    # or a mock for testing.
    yield OpenAIAnalyzer()
