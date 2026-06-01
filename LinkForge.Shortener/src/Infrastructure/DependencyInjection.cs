using Application.Abstractions;
using Infrastructure.Cache;
using Infrastructure.Messaging;
using Infrastructure.Persistence;
using Infrastructure.Persistence.Repositories;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        services.AddDbContext<AppDbContext>(options => options.UseNpgsql(configuration.GetConnectionString("DefaultConnection")));
        services.AddScoped<IShortLinkRepository, ShortLinkRepository>();
        services.AddStackExchangeRedisCache(options =>
        {
            options.Configuration = configuration.GetConnectionString("Redis");
        });
        services.AddScoped<ILinkCache, LinkCache>();

        services.Configure<KafkaSettings>(configuration.GetSection("Kafka"));
        var kafkaEnabled = configuration.GetValue<bool?>("Kafka:Enabled") ?? true;
        if (kafkaEnabled)
        {
            services.AddSingleton<IEventPublisher, KafkaEventPublisher>();
        }
        else
        {
            services.AddSingleton<IEventPublisher, NoopEventPublisher>();
        }

        return services;
    }
}