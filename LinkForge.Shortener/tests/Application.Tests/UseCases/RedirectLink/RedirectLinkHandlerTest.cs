using Application.Abstractions;
using Application.UseCases.RedirectLink;
using Domain.ShortLink;
using FluentAssertions;
using Microsoft.Extensions.Logging.Abstractions;
using NSubstitute;

namespace Application.Tests.UseCases.RedirectLink;


public class RedirectLinkHandlerTests
{
    private readonly IShortLinkRepository _repository;
    private readonly RedirectLinkHandler _handler;
    private readonly ILinkCache _cache;

    public RedirectLinkHandlerTests()
    {
        _repository = Substitute.For<IShortLinkRepository>();
        _cache = Substitute.For<ILinkCache>();
        _handler = new RedirectLinkHandler(_repository, _cache, NullLogger<RedirectLinkHandler>.Instance);
    }

    [Fact]
    public async Task Should_Return_Url_When_ShortCode_Exists()
    {
        var shortLink = ShortLink.Create(new OriginalUrl("https://google.com"), new ShortCode("abcd1234"), DateTimeOffset.UtcNow);

        _cache.GetAsync("abcd1234", Arg.Any<CancellationToken>()).Returns((string?)null);
        _repository.GetByShortCodeAsync("abcd1234", Arg.Any<CancellationToken>()).Returns(shortLink);

        var query = new RedirectLinkQuery("abcd1234");

        var result = await _handler.HandleAsync(query);

        result.Should().Be("https://google.com");
    }

    [Fact]
    public async Task Should_Return_Null_When_ShortCode_NotExists()
    {
        _cache.GetAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns((string?)null);
        _repository.GetByShortCodeAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns((ShortLink?)null);

        var query = new RedirectLinkQuery("inexistente");

        var result = await _handler.HandleAsync(query);

        result.Should().BeNull();
    }
}