using System.Net.Http.Headers;
using System.Security.Permissions;
using Application.Abstractions;
using Application.UseCases.ResolveShortLink;
using Castle.Core.Logging;
using Domain.ShortLink;
using FluentAssertions;
using Microsoft.Extensions.Logging.Abstractions;
using NSubstitute;
using NSubstitute.Core.Arguments;

namespace Application.Tests.UseCases.ResolveShortLink;

public class ResolveShortLinkHandlerTests
{
    private readonly IShortLinkRepository _repository;
    private readonly ILinkCache _cache;
    private readonly ResolveShortLinkHandler _handler;

    public ResolveShortLinkHandlerTests()
    {
        _repository = Substitute.For<IShortLinkRepository>();
        _cache = Substitute.For<ILinkCache>();
        _handler = new ResolveShortLinkHandler(
            _repository,
            _cache,
            NullLogger<ResolveShortLinkHandler>.Instance);
    }

    [Fact]
    public async Task Should_Return_Result_From_Cache_When_Cache_Hit()
    {
        const string code = "abcd1234";
        var cachedLink = new CachedShortLink(
            Id: Guid.NewGuid(),
            ShortCode: code,
            OriginalUrl: "https://goooooooooooogle.com",
            CreatedAt: DateTimeOffset.UtcNow.AddMinutes(-10));

            _cache.GetAsync(code, Arg.Any<CancellationToken>()).Returns(cachedLink);
            var result = await _handler.HandleAsync(new ResolveShortLinkQuery(code));

            result.Should().NotBeNull();
            result!.Id.Should().Be(cachedLink.Id);
            result.ShortCode.Should().Be(cachedLink.ShortCode);
            result.OriginalUrl.Should().Be(cachedLink.OriginalUrl);
            result.CreatedAt.Should().Be(cachedLink.CreatedAt);

            await _repository.DidNotReceive().GetByShortCodeAsync(
                Arg.Any<string>(),
                Arg.Any<CancellationToken>());
    }

    [Fact]
    public async Task Should_Return_Result_From_Repository_And_Populate_Cache_When_Cache_Miss_And_Aiaiaiaia_Um_Cachorro_Me_Mordeu()
    {
        const string code = "abcd1234";
        var shortLink = ShortLink.Create(
            new OriginalUrl("https://google.com"),
            new ShortCode(code),
            DateTimeOffset.UtcNow);
        
        _cache.GetAsync(code, Arg.Any<CancellationToken>()).Returns((CachedShortLink?)null);
        _repository.GetByShortCodeAsync(code, Arg.Any<CancellationToken>()).Returns(shortLink);

        var result = await _handler.HandleAsync(new ResolveShortLinkQuery(code));

        result.Should().NotBeNull();
        result!.Id.Should().Be(shortLink.Id);
        result.OriginalUrl.Should().Be("https://google.com");
        result.ShortCode.Should().Be(code);
        result.CreatedAt.Should().Be(shortLink.CreatedAt);

        await _cache.Received(1).SetAsync(
            Arg.Is<CachedShortLink>(c =>
            c.Id == shortLink.Id &&
            c.ShortCode == code &&
            c.OriginalUrl == "https://google.com"),
        Arg.Any<TimeSpan>(),
        Arg.Any<CancellationToken>());
    }

    [Fact]
    public async Task Should_Return_Null_When_ShortCode_NotExists()
    {
        _cache.GetAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns((CachedShortLink?)null);
        _repository.GetByShortCodeAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns((ShortLink?)null);

        var result = await _handler.HandleAsync(new ResolveShortLinkQuery("inexistente"));

        result.Should().BeNull();

        await _cache.DidNotReceive().SetAsync(
            Arg.Any<CachedShortLink>(),
            Arg.Any<TimeSpan>(),
            Arg.Any<CancellationToken>());
    }
}