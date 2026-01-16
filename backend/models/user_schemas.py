from pydantic import BaseModel, EmailStr
from typing import Optional

class UserCreate(BaseModel):
    email: EmailStr
    password: str
    full_name: Optional[str] = None

class UserUpdate(BaseModel):
    full_name: Optional[str] = None
    password: Optional[str] = None
    avatar_url: Optional[str] = None

class UserResponse(BaseModel):
    id: int
    email: EmailStr
    full_name: Optional[str] = None
    github_id: Optional[str] = None
    avatar_url: Optional[str] = None

    class Config:
        from_attributes = True

class Token(BaseModel):
    access_token: str
    token_type: str
