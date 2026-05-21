using Application.Abstractions;
using Microsoft.Extensions.Caching.Distributed;
using System.Text;

namespace Infrastructure.Cache;

public class LinkCache(IDistributedCache cache) : ILinkCache
{
    private static string Key(string shortCode) => $"redirect:{shortCode}";

    public async Task<string?> GetAsync(string shortCode, CancellationToken cancellationToken = default)
    {
        var bytes = await cache.GetAsync(Key(shortCode), cancellationToken);
        return bytes is null ? null : Encoding.UTF8.GetString(bytes);
    }

    public async Task SetAsync(string shortCode, string originalUrl, TimeSpan ttl, CancellationToken cancellationToken = default)
    {
        var bytes = Encoding.UTF8.GetBytes(originalUrl);
        await cache.SetAsync(Key(shortCode), bytes, new DistributedCacheEntryOptions {AbsoluteExpirationRelativeToNow = ttl}, cancellationToken);
    }
}