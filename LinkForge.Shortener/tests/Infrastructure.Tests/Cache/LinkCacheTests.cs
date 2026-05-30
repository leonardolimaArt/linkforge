using System.Text;
using System.Text.Json;
using Application.Abstractions;
using FluentAssertions;
using Infrastructure.Cache;
using Microsoft.Extensions.Caching.Distributed;
using NSubstitute;

namespace Infrastructure.Tests.Cache;

public class LinkCacheTests
{
    private readonly IDistributedCache _distributedCache;
    private readonly LinkCache _cache;

    public LinkCacheTests()
    {
        _distributedCache = Substitute.For<IDistributedCache>();
        _cache = new LinkCache(_distributedCache);
    }

    [Fact]
    public async Task SetAsync_Should_Serialize_To_Json_And_Store_With_Key_Prefix()
    {
        var link = new CachedShortLink(
            Id: Guid.NewGuid(),
            ShortCode: "abcd1234",
            OriginalUrl: "https://google.com",
            CreatedAt: DateTimeOffset.UtcNow);

        await _cache.SetAsync(link, TimeSpan.FromHours(1));
        await _distributedCache.Received(1).SetAsync(
            "redirect:abcd1234",
            Arg.Is<byte[]>(b => IsJsonContaining(b, "abcd1234") && IsJsonContaining(b, "google.com")),
            Arg.Any<DistributedCacheEntryOptions>(),
            Arg.Any<CancellationToken>());
    }

    [Fact]
    public async Task GetAsync_Should_Return_Object_When_Cache_Hit_With_Valid_Json()
    {
        var stored = new CachedShortLink(
            Id: Guid.NewGuid(),
            ShortCode: "abcd1234",
            OriginalUrl: "https://google.com",
            CreatedAt: DateTimeOffset.UtcNow);

        var bytes = JsonSerializer.SerializeToUtf8Bytes(stored, new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase
        });

        _distributedCache.GetAsync("redirect:abcd1234", Arg.Any<CancellationToken>())
            .Returns(bytes);
        var result = await _cache.GetAsync("abcd1234");

        result.Should().NotBeNull();
        result!.Id.Should().Be(stored.Id);
        result.ShortCode.Should().Be(stored.ShortCode);
        result.OriginalUrl.Should().Be(stored.OriginalUrl);
    }

    [Fact]
    public async Task GetAsync_Should_Return_Null_When_Cache_Miss()
    {
        _distributedCache.GetAsync(Arg.Any<string>(), Arg.Any<CancellationToken>())
            .Returns((byte[]?)null);

        var result = await _cache.GetAsync("missing");

        result.Should().BeNull();
    }

    [Fact]
    public async Task GetAsync_Should_Return_Null_When_Legacy_String_Entry_InCache()
    {
        var legacyBytes = Encoding.UTF8.GetBytes("https://google.com");

        _distributedCache.GetAsync(Arg.Any<string>(), Arg.Any<CancellationToken>())
            .Returns(legacyBytes);

        var result = await _cache.GetAsync("abcd1234");
        result.Should().BeNull();
    }

    private static bool IsJsonContaining(byte[] bytes, string substring) => Encoding.UTF8.GetString(bytes).Contains(substring);
}