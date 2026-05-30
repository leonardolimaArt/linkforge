using Api.Grpc;
using Application.Abstractions;
using Application.UseCases.ResolveShortLink;
using Domain.ShortLink;
using FluentAssertions;
using Grpc.Core;
using LinkForge.Grpc.V1;
using Microsoft.Extensions.Logging.Abstractions;
using NSubstitute;

namespace Api.Tests.Grpc;

public class LinkServiceImplTests
{
    private readonly IShortLinkRepository _repository;
    private readonly ILinkCache _cache;
    private readonly LinkServiceImpl _service;

    public LinkServiceImplTests()
    {
        _repository = Substitute.For<IShortLinkRepository>();
        _cache = Substitute.For<ILinkCache>();

        var handler = new ResolveShortLinkHandler(
            _repository,
            _cache,
            NullLogger<ResolveShortLinkHandler>.Instance);
    
        _service = new LinkServiceImpl(handler, NullLogger<LinkServiceImpl>.Instance);
    }

    [Theory]
    [InlineData("")]
    [InlineData(" ")]
    public async Task Resolve_Should_Throw_InvalidArgument_When_ShortCode_Blank(string shortCode)
    {
        var request = new ResolveRequest{ ShortCode = shortCode};
        var context = TestServerCallContext.Create();
        var result = async () => await _service.Resolve(request, context);

        var except = await result.Should().ThrowAsync<RpcException>();
        except.Which.StatusCode.Should().Be(StatusCode.InvalidArgument);
    }

    [Fact]
    public async Task Resolve_Should_Throw_NotFound_When_ShortCode_Does_Not_Exists()
    {
        const string code = "missing";
        _cache.GetAsync(code, Arg.Any<CancellationToken>()).Returns((CachedShortLink?)null);
        _repository.GetByShortCodeAsync(code, Arg.Any<CancellationToken>()).Returns((ShortLink?)null);

        var request = new ResolveRequest {ShortCode = code};
        var context =  TestServerCallContext.Create();

        var result = async () => await _service.Resolve(request, context);

        var except = await result.Should().ThrowAsync<RpcException>();
        except.Which.StatusCode.Should().Be(StatusCode.NotFound);
    }

    [Fact]
    public async Task Resolve_Should_Return_Response_When_Found_In_Postrgres()
    {
        const string code = "abcd1234";
        var shortLink = ShortLink.Create(
            new OriginalUrl("https://googlehahahah.com/from-db"),
            new ShortCode(code),
            DateTimeOffset.UtcNow
        );

        _cache.GetAsync(code, Arg.Any<CancellationToken>()).Returns((CachedShortLink?)null);
        _repository.GetByShortCodeAsync(code, Arg.Any<CancellationToken>()).Returns(shortLink);

        var request = new ResolveRequest {ShortCode = code};
        var context =  TestServerCallContext.Create();
        var response = await _service.Resolve(request, context);

        response.Id.Should().Be(shortLink.Id.ToString());
        response.ShortCode.Should().Be(code);
        response.OriginalUrl.Should().Be("https://googlehahahah.com/from-db");
        response.CreatedAt.Should().NotBeNull();
    }

    [Fact]
    public async Task Resolve_Should_Return_Response_From_Cache_Without_Hitting_Repository()
    {
        const string code = "abcd1234";
        var cached = new CachedShortLink(
            Id: Guid.NewGuid(),
            ShortCode: code,
            OriginalUrl: "https://google.com/from-cache",
            CreatedAt: DateTimeOffset.UtcNow.AddMinutes(-15)
        );

        _cache.GetAsync(code, Arg.Any<CancellationToken>()).Returns(cached);

        var request = new ResolveRequest {ShortCode = code};
        var context = TestServerCallContext.Create();

        var response = await _service.Resolve(request, context);

        response.Id.Should().Be(cached.Id.ToString());
        response.OriginalUrl.Should().Be("https://google.com/from-cache");

        await _repository.DidNotReceive().GetByShortCodeAsync(
            Arg.Any<string>(),
            Arg.Any<CancellationToken>()
        );
    }

}