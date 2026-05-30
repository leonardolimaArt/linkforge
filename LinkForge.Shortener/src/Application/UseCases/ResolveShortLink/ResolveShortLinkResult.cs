namespace Application.UseCases.ResolveShortLink;

public record ResolveShortLinkResult(
    Guid Id,
    string ShortCode,
    string OriginalUrl,
    DateTimeOffset CreatedAt
);