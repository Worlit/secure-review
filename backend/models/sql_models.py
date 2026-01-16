from sqlalchemy import ForeignKey, String, Integer, DateTime, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship
from datetime import datetime
from typing import List, Optional
from backend.core.database import Base


class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String, unique=True, index=True)
    hashed_password: Mapped[Optional[str]] = mapped_column(String, nullable=True) # For self-registration
    full_name: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    github_id: Mapped[Optional[str]] = mapped_column(String, nullable=True) # For GitHub OAuth
    avatar_url: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    created_at: Mapped[datetime] = mapped_column(default=datetime.utcnow)

    projects: Mapped[List["Project"]] = relationship(back_populates="user")


class Project(Base):
    __tablename__ = "projects"

    id: Mapped[int] = mapped_column(primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    name: Mapped[str] = mapped_column(String)
    repo_url: Mapped[Optional[str]] = mapped_column(String, nullable=True)

    user: Mapped["User"] = relationship(back_populates="projects")
    analyses: Mapped[List["Analysis"]] = relationship(back_populates="project")


class Analysis(Base):
    __tablename__ = "analyses"

    id: Mapped[int] = mapped_column(primary_key=True)
    project_id: Mapped[int] = mapped_column(ForeignKey("projects.id"))
    status: Mapped[str] = mapped_column(String)  # e.g., "pending", "completed", "failed"
    created_at: Mapped[datetime] = mapped_column(default=datetime.utcnow)

    project: Mapped["Project"] = relationship(back_populates="analyses")
    vulnerabilities: Mapped[List["Vulnerability"]] = relationship(back_populates="analysis")


class Vulnerability(Base):
    __tablename__ = "vulnerabilities"

    id: Mapped[int] = mapped_column(primary_key=True)
    analysis_id: Mapped[int] = mapped_column(ForeignKey("analyses.id"))
    type: Mapped[str] = mapped_column(String)  # e.g. SQL Injection
    severity: Mapped[str] = mapped_column(String)  # e.g. High
    file: Mapped[str] = mapped_column(String)
    line: Mapped[str] = mapped_column(String)  # Can be a range or single line number string
    description: Mapped[str] = mapped_column(Text)

    analysis: Mapped["Analysis"] = relationship(back_populates="vulnerabilities")
