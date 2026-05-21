namespace Application.Abstractions;

public interface ILinkCache
{
    Task<string?> GetAsync(string shortCode, CancellationToken cancellationToken = default);
    Task SetAsync(string shortCode, string originalUrl, TimeSpan ttl, CancellationToken cancellationToken = default);
}