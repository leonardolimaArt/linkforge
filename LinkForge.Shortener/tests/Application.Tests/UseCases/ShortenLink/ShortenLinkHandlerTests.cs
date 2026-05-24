using Application.Abstractions;
using Application.UseCases.ShortenLink;
using Domain.ShortLink;
using FluentAssertions;
using Microsoft.Extensions.Logging;
using NSubstitute;

namespace Application.Tests.UseCases.ShortenLink;

public class ShortenLinkHandlerTests
{
    private readonly IShortLinkRepository _repository;
    private readonly ShortenLinkHandler _handler;
    private readonly ILinkCache _cache;
    private readonly IEventPublisher _publisher;
    private readonly ILogger<ShortenLinkHandler> _logger;

    public ShortenLinkHandlerTests()
    {
        _repository = Substitute.For<IShortLinkRepository>();
        _cache = Substitute.For<ILinkCache>();
        _publisher = Substitute.For<IEventPublisher>();
        _logger = Substitute.For<ILogger<ShortenLinkHandler>>();
        _handler = new ShortenLinkHandler(_repository, _cache, _publisher, _logger);
    }

    [Fact]
    public async Task Should_Return_ShortCode_When_Valid_Url()
    {
        var command = new ShortenLinkCommand("https://google.com");
        var result = await _handler.HandleAsync(command);

        result.Should().NotBeNullOrWhiteSpace();
        result.Length.Should().Be(8);
    }
    [Fact]
    public async Task Should_Save_ShortLink_In_Repository()
    {
        var command = new ShortenLinkCommand("https://google.com");
        
        await _handler.HandleAsync(command);
        await _repository.Received(1).AddAsync(Arg.Any<ShortLink>(), Arg.Any<CancellationToken>());
        
    }

    [Fact]
    public async Task Should_Throw_Exception_When_Invalid_Url()
    {
        var command = new ShortenLinkCommand("url-invalida");
        var action = async () => await _handler.HandleAsync(command);
        await action.Should().ThrowAsync<ArgumentException>();
    }
}