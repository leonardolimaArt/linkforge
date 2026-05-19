# LINKFORGE [Handcrafted]

A URL shortener REST API built with .NET 10 and PostgreSQL

## Stack
- .NET 10 / ASP.NET Core
- Entity Framework Core + PostgreSQL
- Docker + Docker Compose
- xUnit, FluentAssertions, NSubstitute, Testcontainers

## Architecture
Clean Architecture with four layers: Domain, Application, Infrastructure, and Api
Value objects (`ShortCode`, `OriginalUrl`) enforce domain rules at the type level
Repository pattern decouples the domain from persistence

## Endpoints
POST - `/api/links` - Shorten a URL
GET - `/r/{shortCode}` - Redirect to original URL

## Running locally

Requires Docker

```bash
git clone https://github.com/leonardolimaArt/linkforge.git
cd linkforge/LinkForge.Shortener
cp .env.example .env
docker compose up --build
API runs on http://localhost:8080

Tests

dotnet test LinkForge.Shortener/tests/Domain.Tests
dotnet test LinkForge.Shortener/tests/Application.Tests
dotnet test LinkForge.Shortener/tests/Infrastructure.Tests
dotnet test LinkForge.Shortener/tests/Api.Tests
Integration and functional tests require Docker, Testcontainers handles that automatically

Environments
Environment	Branch
Production	main
Development	develop
Deploys are triggered automatically via GitHub Actions after all tests pass