using Application.UseCases.RedirectLink;
using Application.UseCases.ShortenLink;
using Microsoft.Extensions.DependencyInjection;

namespace Application;


public static class DependencyInjection
{
    public static IServiceCollection AddApplication(this IServiceCollection services)
    {
        services.AddScoped<ShortenLinkHandler>();
        services.AddScoped<RedirectLinkHandler>();

        return services;
    }
}