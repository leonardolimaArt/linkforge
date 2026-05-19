using Infrastructure.Persistence;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Api.Tests.Fixtures;

public class ApiFixture : IAsyncLifetime
{
    public CustomWebApplicationFactory Factory {get;} = new();
    public HttpClient Client {get; private set;} = null!;

    public async Task InitializeAsync()
    {
        await Factory.InitializeAsync();

        using var scope = Factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        await db.Database.MigrateAsync();

        Client = Factory.CreateClient(new WebApplicationFactoryClientOptions
        {
            AllowAutoRedirect = false
        });
    }

    public async Task DisposeAsync()
    {
        using var scope = Factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        db.ShortLinks.RemoveRange(db.ShortLinks);
        await db.SaveChangesAsync();

        await Factory.DisposeAsync();
    }
}