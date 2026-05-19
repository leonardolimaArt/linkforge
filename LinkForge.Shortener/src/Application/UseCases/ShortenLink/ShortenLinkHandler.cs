using Application.Abstractions;
using Domain.ShortLink;

namespace Application.UseCases.ShortenLink;

public class ShortenLinkHandler
{
    private readonly IShortLinkRepository _repository;

    public ShortenLinkHandler(IShortLinkRepository repository)
    {
        _repository = repository;
    }

    public async Task<string> HandleAsync(ShortenLinkCommand command, CancellationToken cancellationToken = default)
    {
        var originalUrl = new OriginalUrl(command.OriginalUrl);
        var shortCode = new ShortCode(GenerateCode());
        var shortLink = ShortLink.Create(originalUrl, shortCode, DateTimeOffset.UtcNow);

        await _repository.AddAsync(shortLink, cancellationToken);

        return shortCode.Value;
    }

    private static string GenerateCode()
    {
        return  Guid.NewGuid().ToString("N")[..8];
    }
}