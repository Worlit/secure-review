from pydantic_settings import BaseSettings, SettingsConfigDict
from functools import lru_cache
import os


class Settings(BaseSettings):
    PROJECT_NAME: str = "Secure Review"
    API_V1_STR: str = "/api/v1"
    OPENAI_API_KEY: str = ""
    DATABASE_URL: str

    # GitHub OAuth (Optional placeholders)
    GITHUB_CLIENT_ID: str = ""
    GITHUB_CLIENT_SECRET: str = ""

    model_config = SettingsConfigDict(
        env_file=os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), ".env"),
        env_file_encoding="utf-8",
        extra="ignore"
    )


@lru_cache
def get_settings():
    return Settings()
