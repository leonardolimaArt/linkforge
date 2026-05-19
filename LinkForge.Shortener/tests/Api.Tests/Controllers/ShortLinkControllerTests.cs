using System.Net;
using System.Net.Http.Json;
using Api.Tests.Fixtures;
using Application.UseCases.ShortenLink;
using Domain.ShortLink;
using FluentAssertions;

namespace Api.Tests.Controllers;

public class ShortLinkControllerTests : IClassFixture<ApiFixture>
{
    private readonly HttpClient _client;

    public ShortLinkControllerTests(ApiFixture apiFixture)
    {
        _client = apiFixture.Client;
    }

    [Fact]
    public async Task POST_Should_Return200_AndShortCode_When_Url_IsValid()
    {
        var request = new ShortenLinkRequest("https://google.com");

        var response = await _client.PostAsJsonAsync("/api/links", request);

        response.StatusCode.Should().Be(HttpStatusCode.OK);

        var body = await response.Content.ReadFromJsonAsync<ShortenLinkResponse>();
        body.Should().NotBeNull();
        body!.shortCode.Should().NotBeNullOrEmpty();
    }

    [Fact]
    public async Task GET_Should_Redirect_When_ShortCode_Exists()
    {
        var request = new ShortenLinkRequest("https://google.com");
        var postResponse = await _client.PostAsJsonAsync("/api/links", request);
        var body = await postResponse.Content.ReadFromJsonAsync<ShortenLinkResponse>();

        var response = await _client.GetAsync($"/r/{body!.shortCode}");

        response.StatusCode.Should().Be(HttpStatusCode.Redirect);
    }

    [Fact]
    public async Task GET_ShouldReturn404_WhenShortCodeNotExists()
    {
        var response = await _client.GetAsync("/r/inexistente");
        response.StatusCode.Should().Be(HttpStatusCode.NotFound);
    }

    public record ShortenLinkResponse(string shortCode);
}