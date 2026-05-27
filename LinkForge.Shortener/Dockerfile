FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /source

COPY proto/ proto/
COPY LinkForge.Shortener/LinkForge.Shortener.slnx LinkForge.Shortener/
COPY LinkForge.Shortener/src/Api/Api.csproj LinkForge.Shortener/src/Api/
COPY LinkForge.Shortener/src/Application/Application.csproj LinkForge.Shortener/src/Application/
COPY LinkForge.Shortener/src/Domain/Domain.csproj LinkForge.Shortener/src/Domain/
COPY LinkForge.Shortener/src/Infrastructure/Infrastructure.csproj LinkForge.Shortener/src/Infrastructure/

RUN dotnet restore LinkForge.Shortener/src/Api/Api.csproj

COPY LinkForge.Shortener/src/ LinkForge.Shortener/src/

RUN dotnet publish LinkForge.Shortener/src/Api/Api.csproj -c Release -o /app

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS runtime
WORKDIR /app

COPY --from=build /app .

EXPOSE 8080
EXPOSE 5001

ENTRYPOINT ["dotnet", "Api.dll"]