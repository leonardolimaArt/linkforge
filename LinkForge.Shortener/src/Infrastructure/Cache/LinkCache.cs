using Application.Abstractions;
using Microsoft.Extensions.Caching.Distributed;
using System.Text;
using System.Text.Json;

namespace Infrastructure.Cache;

public class LinkCache(IDistributedCache cache) : ILinkCache
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase
    };
    private static string Key(string shortCode) => $"redirect:{shortCode}";

    public async Task<CachedShortLink?> GetAsync(string shortCode, CancellationToken cancellationToken = default)
    {
        var bytes = await cache.GetAsync(Key(shortCode), cancellationToken);
        if (bytes is null || bytes.Length == 0)
        {
            return null;
        }

        try
        {
            return JsonSerializer.Deserialize<CachedShortLink>(bytes, JsonOptions);
        }
        catch (JsonException)
        {
            return null;
        }
    }

    public async Task SetAsync(CachedShortLink link, TimeSpan ttl, CancellationToken cancellationToken = default)
    {
        var bytes = JsonSerializer.SerializeToUtf8Bytes(link, JsonOptions);
        var options = new DistributedCacheEntryOptions
        {
            AbsoluteExpirationRelativeToNow = ttl
        };
        await cache.SetAsync(Key(shortCode: link.ShortCode), bytes, options, cancellationToken);
    }
}