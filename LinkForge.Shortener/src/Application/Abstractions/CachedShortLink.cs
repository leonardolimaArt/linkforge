namespace Application.Abstractions;

public record CachedShortLink(
    Guid Id,
    string ShortCode,
    string OriginalUrl,
    DateTimeOffset CreatedAt
);