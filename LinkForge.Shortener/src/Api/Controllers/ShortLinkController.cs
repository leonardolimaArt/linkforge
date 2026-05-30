using Application.UseCases.ResolveShortLink;
using Application.UseCases.ShortenLink;
using Microsoft.AspNetCore.Mvc;

namespace Api.Controllers;

[ApiController]
[Route("api/links")]
public class ShortLinkController : ControllerBase
{
    private readonly ShortenLinkHandler _shortenHandler;
    private readonly ResolveShortLinkHandler _resolveHandler;

    public ShortLinkController(ShortenLinkHandler shortenHandler,  ResolveShortLinkHandler resolveHandler)
    {
        _resolveHandler = resolveHandler;
        _shortenHandler = shortenHandler;
    }

    [HttpPost]
    public async Task<IActionResult> ShortenLink([FromBody] ShortenLinkRequest request, CancellationToken cancellationToken)
    {
        var shortCode = await _shortenHandler.HandleAsync(new ShortenLinkCommand(request.Url), cancellationToken);
        
        return Ok(new {shortCode});
    }

    [HttpGet("/r/{shortCode}")]
    public async Task<IActionResult> RedirectLink([FromRoute] string shortCode, CancellationToken cancellationToken)
    {
        var url = await _resolveHandler.HandleAsync(new ResolveShortLinkQuery(shortCode), cancellationToken);

        if (url is null)
            return NotFound();

        return Redirect(url.OriginalUrl);
    }
}