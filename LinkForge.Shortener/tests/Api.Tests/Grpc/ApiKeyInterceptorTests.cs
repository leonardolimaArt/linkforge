using Api.Grpc;
using FluentAssertions;
using Grpc.Core;
using Grpc.Core.Interceptors;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging.Abstractions;
using NSubstitute;

namespace Api.Tests.Grpc;

public class ApiKeyInterceptorTests
{
    private const string ConfiguredKey = "minha-key-bem-legal-hahahhahahahahahahahahahahahaahahahahahahahaaha";

    private readonly IConfiguration _config;
    private readonly ApiKeyInterceptor _interceptor;

    public ApiKeyInterceptorTests()
    {
        _config = Substitute.For<IConfiguration>();
        _interceptor = new ApiKeyInterceptor(_config, NullLogger<ApiKeyInterceptor>.Instance);
    }

    [Fact]
    public async Task Should_Thow_Unautenticated_when_Server_Has_No_Api_Key_Configured()
    {
        _config["API_KEY"].Returns((string?)null);
        var context = TestServerCallContext.Create();
        UnaryServerMethod<string, string> continuation = (_, _) => Task.FromResult("ok");

        var result = async () => await _interceptor.UnaryServerHandler("req", context, continuation);
        var except = await result.Should().ThrowAsync<RpcException>();
        except.Which.StatusCode.Should().Be(StatusCode.Unauthenticated);
    }

    [Fact]
    public async Task Should_Throw_Unauthenticated_When_ApiKey_Header_Missing()
    {
        _config["API_KEY"].Returns(ConfiguredKey);
        var context = TestServerCallContext.Create();
        UnaryServerMethod<string, string> continuation = (_, _) => Task.FromResult("ok");

        var result = async () => await _interceptor.UnaryServerHandler("req", context, continuation);
        var except = await result.Should().ThrowAsync<RpcException>();
        except.Which.StatusCode.Should().Be(StatusCode.Unauthenticated);
    }

    [Fact]
    public async Task Should_Throw_Unauthenticated_When_ApiKey_Header_Wrong()
    {
        _config["API_KEY"].Returns(ConfiguredKey);
        var metadata = new Metadata { {"x-api-key", "key"}};
        var context = TestServerCallContext.Create(metadata);
        UnaryServerMethod<string, string> continuation = (_, _) => Task.FromResult("ok");

        var result = async () => await _interceptor.UnaryServerHandler("req", context, continuation);
        var except = await result.Should().ThrowAsync<RpcException>();
        except.Which.StatusCode.Should().Be(StatusCode.Unauthenticated);
    }

    [Fact]
    public async Task Should_Call_Continuation_When_ApiKey_Matches()
    {
        _config["API_KEY"].Returns(ConfiguredKey);
        var metadata = new Metadata { {"x-api-key", ConfiguredKey}};
        var context = TestServerCallContext.Create(metadata);

        var continuationCalled = false;
        UnaryServerMethod<string, string> continuation = (_, _) => {
            continuationCalled = true;
            return Task.FromResult("ok");
        };

        var result = await _interceptor.UnaryServerHandler("req", context, continuation);
        continuationCalled.Should().BeTrue();
        result.Should().Be("ok");
    }
}