using Domain.ShortLink;

namespace Application.Abstractions;

public interface IShortLinkRepository
{
    Task AddAsync(ShortLink shortLink, CancellationToken cancellationToken = default);
    Task<ShortLink?> GetByShortCodeAsync(string shortCode, CancellationToken cancellationToken = default);
}