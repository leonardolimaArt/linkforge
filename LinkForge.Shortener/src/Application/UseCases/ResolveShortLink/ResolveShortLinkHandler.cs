using Application.Abstractions;
using Microsoft.Extensions.Logging;

namespace Application.UseCases.ResolveShortLink;

public class ResolveShortLinkHandler(IShortLinkRepository repository, ILinkCache cache, ILogger<ResolveShortLinkHandler> logger)
{
    private static readonly TimeSpan CacheTtl = TimeSpan.FromHours(1);
    public async Task<ResolveShortLinkResult?> HandleAsync(ResolveShortLinkQuery query, CancellationToken cancellationToken = default)
    {
        var cachedUrl = await cache.GetAsync(query.ShortCode, cancellationToken);
        if (cachedUrl is not null)
        {
            logger.LogInformation("Cache HIT for {ShortCode}", query.ShortCode);
            return new ResolveShortLinkResult(
                Id: cachedUrl.Id,
                ShortCode: cachedUrl.ShortCode,
                OriginalUrl: cachedUrl.OriginalUrl,
                CreatedAt: cachedUrl.CreatedAt);
        }

        logger.LogInformation("Cahce MISS for {ShortCode}", query.ShortCode);
        var shortLink = await repository.GetByShortCodeAsync(query.ShortCode, cancellationToken);
        if(shortLink is null)
            return null;

        await cache.SetAsync(new CachedShortLink(Id: shortLink.Id, ShortCode: shortLink.ShortCode.Value, OriginalUrl: shortLink.OriginalUrl.Value, CreatedAt: shortLink.CreatedAt), CacheTtl, cancellationToken);

        return new ResolveShortLinkResult(
            Id: shortLink.Id,
            ShortCode: shortLink.ShortCode.Value,
            OriginalUrl: shortLink.OriginalUrl.Value,
            CreatedAt: shortLink.CreatedAt
            );
    }
}