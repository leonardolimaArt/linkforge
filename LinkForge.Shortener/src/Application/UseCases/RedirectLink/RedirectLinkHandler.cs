using Application.Abstractions;

namespace Application.UseCases.RedirectLink;

public class RedirectLinkHandler(IShortLinkRepository repository, ILinkCache cache)
{
    public async Task<string?> HandleAsync(RedirectLinkQuery query, CancellationToken cancellationToken = default)
    {
        var cached = await cache.GetAsync(query.ShortCode, cancellationToken);
        if(cached is not null)
            return cached;

        var shortLink = await repository.GetByShortCodeAsync(query.ShortCode, cancellationToken);
        if(shortLink is null)
            return null;

        await cache.SetAsync(query.ShortCode, shortLink.OriginalUrl.Value, TimeSpan.FromHours(1), cancellationToken);
        return shortLink.OriginalUrl.Value;
    }
}