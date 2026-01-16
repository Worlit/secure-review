from pydantic import BaseModel
from typing import List, Optional

class CodeAnalysisRequest(BaseModel):
    code: str
    language: str
    filename: Optional[str] = None

class VulnerabilityIssue(BaseModel):
    type: str
    location: str
    description: str
    severity: str  # e.g., High, Medium, Low
    recommendation: str
    fix_example: str

class AnalysisResult(BaseModel):
    issues: List[VulnerabilityIssue]
    summary: str
    security_score: int  # 0-100
