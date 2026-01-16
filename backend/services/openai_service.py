import json
from openai import AsyncOpenAI
from backend.services.interfaces import ICodeAnalyzer
from backend.models.schemas import CodeAnalysisRequest, AnalysisResult, VulnerabilityIssue
from backend.core.config import get_settings

settings = get_settings()


class OpenAIAnalyzer(ICodeAnalyzer):
    def __init__(self):
        # We initialize the client here.
        # In a more complex setup, we could inject the client itself.
        self.client = AsyncOpenAI(api_key=settings.OPENAI_API_KEY)

    async def analyze_code(self, request: CodeAnalysisRequest) -> AnalysisResult:
        prompt = self._build_prompt(request)

        try:
            response = await self.client.chat.completions.create(
                model="gpt-4o",  # Or gpt-3.5-turbo depending on budget
                messages=[
                    {
                        "role": "system",
                        "content": "You are a senior cybersecurity expert specializing in SAST (Static Application Security Testing).",
                    },
                    {"role": "user", "content": prompt},
                ],
                response_format={"type": "json_object"},  # Force JSON output
            )

            content = response.choices[0].message.content
            if content is None:
                return AnalysisResult(
                    issues=[],
                    summary="Analysis failed: Empty response from AI",
                    security_score=0,
                )
            return self._parse_response(content)

        except Exception as e:
            # In a real app, log this error properly
            print(f"Error calling OpenAI: {e}")
            # Return a fallback or raise HTTP exception
            return AnalysisResult(issues=[], summary=f"Analysis failed: {str(e)}", security_score=0)

    def _build_prompt(self, request: CodeAnalysisRequest) -> str:
        return f"""
        Analyze the following {request.language} code for OWASP Top 10 vulnerabilities and architectural flaws.
        
        Code:
        ```{request.language}
        {request.code}
        ```
        
        Return the result strictly in this JSON format:
        {{
            "issues": [
                {{
                    "type": "Vulnerability Type",
                    "location": "Line numbers or context",
                    "description": "Why is this dangerous?",
                    "severity": "High/Medium/Low",
                    "recommendation": "How to fix",
                    "fix_example": "Code snippet"
                }}
            ],
            "summary": "Brief executive summary of findings",
            "security_score": 0-100 (integer, where 100 is perfectly secure)
        }}
        """

    def _parse_response(self, content: str) -> AnalysisResult:
        try:
            data = json.loads(content)
            return AnalysisResult(**data)
        except (json.JSONDecodeError, Exception):
            return AnalysisResult(
                issues=[],
                summary="Failed to parse AI response or validation error",
                security_score=0,
            )
