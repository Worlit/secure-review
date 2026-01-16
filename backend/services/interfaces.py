from abc import ABC, abstractmethod
from backend.models.schemas import CodeAnalysisRequest, AnalysisResult

class ICodeAnalyzer(ABC):
    """
    Interface for Code Analyzer (DIP - Dependency Inversion Principle).
    We depend on this abstraction, not concrete OpenAI implementation.
    """
    @abstractmethod
    async def analyze_code(self, request: CodeAnalysisRequest) -> AnalysisResult:
        pass
