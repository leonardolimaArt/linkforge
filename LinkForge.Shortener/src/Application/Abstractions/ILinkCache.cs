namespace Application.Abstractions;

public interface ILinkCache
{
    Task<CachedShortLink?> GetAsync(string shortCode, CancellationToken cancellationToken = default);
    Task SetAsync(CachedShortLink link, TimeSpan ttl, CancellationToken cancellationToken = default);
}