from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import RedirectResponse
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from typing import Optional
import httpx

from backend.core.database import get_db
from backend.core.config import get_settings
from backend.core.security import get_password_hash, verify_password
from backend.models.sql_models import User
from backend.models.user_schemas import UserCreate, UserResponse, Token

router = APIRouter()
settings = get_settings()


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


@router.get("/github/login")
async def github_login():
    return RedirectResponse(
        f"https://github.com/login/oauth/authorize?client_id={settings.GITHUB_CLIENT_ID}&scope=user:email"
    )


@router.get("/github/callback")
async def github_callback(code: str, db: AsyncSession = Depends(get_db)):
    async with httpx.AsyncClient() as client:
        # 1. Exchange code for access token
        response = await client.post(
            "https://github.com/login/oauth/access_token",
            headers={"Accept": "application/json"},
            data={
                "client_id": settings.GITHUB_CLIENT_ID,
                "client_secret": settings.GITHUB_CLIENT_SECRET,
                "code": code,
            },
        )
        response.raise_for_status()
        token_data = response.json()
        access_token = token_data.get("access_token")

        if not access_token:
            error_desc = token_data.get("error_description", "Unknown error")
            raise HTTPException(
                status_code=400, detail=f"Failed to get access token from GitHub: {error_desc}"
            )

        # 2. Get user info
        user_response = await client.get(
            "https://api.github.com/user",
            headers={
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/json",
            },
        )
        user_response.raise_for_status()
        github_user_data = user_response.json()

        # 3. Get user email (if private)
        email_response = await client.get(
            "https://api.github.com/user/emails",
            headers={
                "Authorization": f"Bearer {access_token}",
                "Accept": "application/json",
            },
        )
        email_response.raise_for_status()
        emails = email_response.json()
        primary_email = next((e["email"] for e in emails if e["primary"]), None)

        if not primary_email:
            raise HTTPException(
                status_code=400, detail="No primary email found in GitHub account"
            )
        
        # 4. Find or Create User
        result = await db.execute(select(User).where(User.email == primary_email))
        user = result.scalars().first()

        if not user:
            # Create new user if not exists
            user = User(
                email=primary_email,
                github_id=str(github_user_data["id"]),
                avatar_url=github_user_data.get("avatar_url"),
                full_name=github_user_data.get("name"),
            )
            db.add(user)
            await db.commit()
            await db.refresh(user)
        else:
            # Update github_id if missing (connect account)
            if not user.github_id:
                user.github_id = str(github_user_data["id"])
                user.avatar_url = github_user_data.get("avatar_url") or user.avatar_url
                await db.commit()

        # In a real app, you would generate a JWT here and redirect to frontend with it
        # return RedirectResponse(f"http://localhost:5173/auth/callback?token={jwt_token}")
        
        return {
            "access_token": "fake-jwt-token-from-github-login",
            "token_type": "bearer",
            "user": {"email": user.email, "id": user.id}
        }
