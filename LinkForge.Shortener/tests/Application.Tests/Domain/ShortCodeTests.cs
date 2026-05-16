using Domain.ShortLink;
using FluentAssertions;

namespace Application.Tests.Domain;

public class ShortCodeTests
{
    
    [Fact]
    public void Should_Create_Valid_ShortCode()
    {
        var value = "abc12345";
        var shortCode = new ShortCode(value);
        shortCode.Value.Should().Be(value);
    }

    [Theory]
    [InlineData("")]
    [InlineData(" ")]
    [InlineData(null)]
    public void Should_Throw_Exception_When_Value_IsEmpty(string? value)
    {
        var action = () => new ShortCode(value);
        action.Should().Throw<ArgumentException>();
    }

    [Theory]
    [InlineData("aa")]
    [InlineData("a1a")]
    [InlineData("ahahahahahahahahahahahahahahhahahahahahahaahahahahahahahhaahahhahahahahahah010")]
    public void Should_Throw_Exception_When_Invalid_Size(string? value)
    {
        var action = () => new ShortCode(value);
        action.Should().Throw<ArgumentException>();
    }

    [Theory]
    [InlineData("a1b2")]
    [InlineData("12345678910123456789212345678931")]
    public void Should_Accept_Valid_Sizes(string value)
    {
        var action = () => new ShortCode(value);
        action.Should().NotThrow();
    }
}