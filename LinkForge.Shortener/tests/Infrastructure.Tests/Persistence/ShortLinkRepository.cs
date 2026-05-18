using Domain.ShortLink;
using FluentAssertions;
using Infrastructure.Persistence.Repositories;
using Infrastructure.Tests.Fixtures;
using Microsoft.EntityFrameworkCore;

namespace Infrastructure.Tests.Persistence;

public class ShortLinkRepositoryTests : IClassFixture<DatabaseFixture>, IAsyncLifetime
{
    private readonly DatabaseFixture _fixture;
    private readonly ShortLinkRepository _repository;

    public ShortLinkRepositoryTests(DatabaseFixture fixture)
    {
        _fixture = fixture;
        _repository = new ShortLinkRepository(fixture.DbContext);
    }

    public Task InitializeAsync() => Task.CompletedTask;

    public async Task DisposeAsync()
    {
        _fixture.DbContext.ShortLinks.RemoveRange(_fixture.DbContext.ShortLinks);
        await _fixture.DbContext.SaveChangesAsync();
    }
    
    [Fact]
    public async Task Should_Save_ShortLink_InDataBase()
    {
        var shortLink = ShortLink.Create(new OriginalUrl("https://google.com"), new ShortCode("abcd1234"), DateTimeOffset.UtcNow);

        await _repository.AddAsync(shortLink);

        var result = await _repository.GetByShortCodeAsync("abcd1234");

        result.Should().NotBeNull();
        result!.ShortCode.Value.Should().Be("abcd1234");
        result!.OriginalUrl.Value.Should().Be("https://google.com");

    }

    [Fact]
    public async Task Should_Return_Null_When_NotExists()
    {
        var result = await _repository.GetByShortCodeAsync("efgh5678");
        result.Should().BeNull();
    }

    [Fact]
    public async Task Should_Reject_Duplicated_ShortCode()
    {
        var shortLink1 = ShortLink.Create(new OriginalUrl("https://google.com"), new ShortCode("abcd1234"), DateTimeOffset.UtcNow);
        var shortLink2 = ShortLink.Create(new OriginalUrl("https://github.com"), new ShortCode("abcd1234"), DateTimeOffset.UtcNow);

        await _repository.AddAsync(shortLink1);

        var action = async () => await _repository.AddAsync(shortLink2);

        await action.Should().ThrowAsync<DbUpdateException>();
    }
}