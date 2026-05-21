using Infrastructure.Persistence;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Caching.Distributed;
using Microsoft.Extensions.DependencyInjection;
using Testcontainers.PostgreSql;

namespace Api.Tests.Fixtures;

public class CustomWebApplicationFactory : WebApplicationFactory<Program>, IAsyncLifetime
{
    private readonly PostgreSqlContainer _container = new PostgreSqlBuilder("postgres:16-alpine").Build();

    public async Task InitializeAsync()
    {
        await _container.StartAsync();
    }

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        builder.ConfigureServices(services =>
        {
            var dbdescriptor = services.SingleOrDefault(
                d => d.ServiceType == typeof(DbContextOptions<AppDbContext>));

            if (dbdescriptor is not null)
                services.Remove(dbdescriptor);

            services.AddDbContext<AppDbContext>(options =>
            options.UseNpgsql(_container.GetConnectionString()));

            var redisDescriptor = services.SingleOrDefault(
                d => d.ServiceType == typeof(IDistributedCache));
            if(redisDescriptor is not null)
                services.Remove(redisDescriptor);
            services.AddDistributedMemoryCache();
        });
    }

    public new async Task DisposeAsync()
    {
        await _container.DisposeAsync();
    }
}