from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from typing import Optional

from backend.core.database import get_db
from backend.core.security import get_password_hash, verify_password
from backend.models.sql_models import User
from backend.models.user_schemas import UserCreate, UserResponse, Token

router = APIRouter()


@router.post("/register", response_model=UserResponse)
async def register(user_in: UserCreate, db: AsyncSession = Depends(get_db)):
    # Check if user exists
    result = await db.execute(select(User).where(User.email == user_in.email))
    existing_user = result.scalars().first()

    if existing_user:
        raise HTTPException(status_code=400, detail="User with this email already exists")

    # Create new user
    new_user = User(
        email=user_in.email,
        hashed_password=get_password_hash(user_in.password),
        full_name=user_in.full_name,
    )

    db.add(new_user)
    await db.commit()
    await db.refresh(new_user)

    return new_user


@router.post("/login/password", response_model=Token)
async def login_password(user_in: UserCreate, db: AsyncSession = Depends(get_db)):
    # Simple login for now, JWT generation usually goes here
    result = await db.execute(select(User).where(User.email == user_in.email))
    user = result.scalars().first()

    if (
        not user
        or not user.hashed_password
        or not verify_password(user_in.password, user.hashed_password)
    ):
        raise HTTPException(status_code=400, detail="Incorrect email or password")

    return {"access_token": "fake-jwt-token-for-now", "token_type": "bearer"}


@router.post("/connect-github")
async def connect_github(github_id: str, current_user_id: int):  # Needs real auth
    # Logic to update user with github_id
    pass
