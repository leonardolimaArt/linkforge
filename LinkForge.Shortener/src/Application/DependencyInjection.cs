using Application.UseCases.ResolveShortLink;
using Application.UseCases.ShortenLink;
using Microsoft.Extensions.DependencyInjection;

namespace Application;


public static class DependencyInjection
{
    public static IServiceCollection AddApplication(this IServiceCollection services)
    {
        services.AddScoped<ShortenLinkHandler>();
        services.AddScoped<ResolveShortLinkHandler>();

        return services;
    }
}