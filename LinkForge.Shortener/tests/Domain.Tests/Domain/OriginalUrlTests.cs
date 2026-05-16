using System.Runtime.InteropServices;
using Domain.ShortLink;
using FluentAssertions;

namespace Domain.Tests.Domain;

public class OriginalUrlTests
{
    [Fact]
    public void Should_Create_Valid_OriginalUrl()
    {
        var value = "https://google.com";
        var url = new OriginalUrl(value);

        url.Value.Should().Be(value);
    }

    [Theory]
    [InlineData("")]
    [InlineData(" ")]
    [InlineData(null)]
    public void Should_Throw_Exception_When_Empty_Value(string? value)
    {
        var action = () => new OriginalUrl(value);
        action.Should().Throw<ArgumentException>();
    }

    [Theory]
    [InlineData("google.com")]
    [InlineData("Leonardo")]
    [InlineData("ftp://google.com")]
    public void Should_Throw_Exception_When_Invalid_Url(string value)
    {
        var action = () => new OriginalUrl(value);
        action.Should().Throw<ArgumentException>();
    }

    [Theory]
    [InlineData("http://google.com")]
    [InlineData("https://google.com")]
    [InlineData("https://google.com/search?q=fish+city")]
    public void Should_Accept_Http_Https(string value)
    {
        var action = () => new OriginalUrl(value);
        action.Should().NotThrow();
    }
}