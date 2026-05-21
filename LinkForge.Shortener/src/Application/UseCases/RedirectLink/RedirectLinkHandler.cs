using Application.Abstractions;
using Microsoft.Extensions.Logging;

namespace Application.UseCases.RedirectLink;

public class RedirectLinkHandler(IShortLinkRepository repository, ILinkCache cache, ILogger<RedirectLinkHandler> logger)
{
    public async Task<string?> HandleAsync(RedirectLinkQuery query, CancellationToken cancellationToken = default)
    {
        var cached = await cache.GetAsync(query.ShortCode, cancellationToken);
        if(cached is not null){
            logger.LogInformation("Cache HIT for {shortCode}", query.ShortCode);
            return cached;
        }

        logger.LogInformation("Cache MISS for {shortCode}", query.ShortCode);
        var shortLink = await repository.GetByShortCodeAsync(query.ShortCode, cancellationToken);
        if(shortLink is null)
            return null;

        await cache.SetAsync(query.ShortCode, shortLink.OriginalUrl.Value, TimeSpan.FromHours(1), cancellationToken);
        return shortLink.OriginalUrl.Value;
    }
}