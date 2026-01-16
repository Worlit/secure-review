from fastapi import APIRouter, Depends, HTTPException
from backend.models.schemas import CodeAnalysisRequest, AnalysisResult
from backend.services.interfaces import ICodeAnalyzer
from backend.api.deps import get_analyzer
from backend.api import auth

router = APIRouter()

router.include_router(auth.router, prefix="/auth", tags=["auth"])


@router.post("/analyze", response_model=AnalysisResult)
async def analyze_code(
    request: CodeAnalysisRequest, analyzer: ICodeAnalyzer = Depends(get_analyzer)
):
    """
    Analyzes the provided code snippet for security vulnerabilities using AI.
    """
    if not request.code.strip():
        raise HTTPException(status_code=400, detail="Code cannot be empty")

    result = await analyzer.analyze_code(request)
    return result
