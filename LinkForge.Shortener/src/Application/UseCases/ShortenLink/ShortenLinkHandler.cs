using Application.Abstractions;
using Application.Events;
using Domain.ShortLink;
using Microsoft.Extensions.Logging;

namespace Application.UseCases.ShortenLink;

public class ShortenLinkHandler
{
    private readonly IShortLinkRepository _repository;
    private const int CurrentSchemaVersion = 1;
    private static readonly TimeSpan CacheTtl1 = TimeSpan.FromHours(1);

    private readonly ILinkCache _cache;
    private readonly IEventPublisher _publisher;
    private readonly ILogger<ShortenLinkHandler> _logger;

    public ShortenLinkHandler(IShortLinkRepository repository, ILinkCache cache, IEventPublisher publisher, ILogger<ShortenLinkHandler> logger)
    {
        _repository = repository;
        _cache = cache;
        _publisher = publisher;
        _logger = logger;
    }

    public async Task<string> HandleAsync(ShortenLinkCommand command, CancellationToken cancellationToken = default)
    {
        var originalUrl = new OriginalUrl(command.OriginalUrl);
        var shortCode = new ShortCode(GenerateCode());
        var shortLink = ShortLink.Create(originalUrl, shortCode, DateTimeOffset.UtcNow);

        await _repository.AddAsync(shortLink, cancellationToken);


        try
        {
            await _cache.SetAsync(new CachedShortLink(Id: shortLink.Id, ShortCode: shortCode.Value, OriginalUrl: originalUrl.Value, CreatedAt: shortLink.CreatedAt), CacheTtl1, cancellationToken);
        }catch(Exception ex)
        {
            _logger.LogWarning(ex, "Cache pre-population failed for {Code}", shortCode.Value);
        }

        _ = _publisher.PublishAsync(
            key: shortCode.Value,
            payload: new LinkCreatedEvent(
                SchemaVersion: CurrentSchemaVersion,
                Id: shortLink.Id,
                ShortCode: shortCode.Value,
                OriginalUrl: originalUrl.Value,
                CreatedAt: shortLink.CreatedAt
            ),
            cancellationToken: CancellationToken.None
        );

        return shortCode.Value;
    }

    private static string GenerateCode()
    {
        return  Guid.NewGuid().ToString("N")[..8];
    }
}