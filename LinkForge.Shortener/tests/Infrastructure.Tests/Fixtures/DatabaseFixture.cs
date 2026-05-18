using Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Testcontainers.PostgreSql;

namespace Infrastructure.Tests.Fixtures;

public class DatabaseFixture : IAsyncLifetime
{
    private readonly PostgreSqlContainer _container = new PostgreSqlBuilder("postgres:16-alpine").Build();

    public AppDbContext DbContext {get; private set; } = null!;

    public async Task InitializeAsync()
    {
        await _container.StartAsync();

        var options =new DbContextOptionsBuilder<AppDbContext>().UseNpgsql(_container.GetConnectionString()).Options;

        DbContext = new AppDbContext(options);

        await DbContext.Database.MigrateAsync();
    }

    public async Task DisposeAsync()
    {
        await DbContext.DisposeAsync();
        await _container.DisposeAsync();
    }
}