using Application.Abstractions;

namespace Application.UseCases.RedirectLink;

public class RedirectLinkHandler
{
    
    private readonly IShortLinkRepository _repository;

    public RedirectLinkHandler(IShortLinkRepository repository)
    {
        _repository = repository;
    }

    public async Task<string?> HandleAsync(RedirectLinkQuery query, CancellationToken cancellationToken = default)
    {
        var shortLink = await _repository.GetByShortCodeAsync(query.ShortCode, cancellationToken);
        return shortLink?.OriginalUrl.Value;
    }
}