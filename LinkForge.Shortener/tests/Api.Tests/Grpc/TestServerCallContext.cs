using Grpc.Core;

namespace Api.Tests.Grpc;

internal class TestServerCallContext : ServerCallContext
{
    private readonly Metadata _requestHeaders;

    private TestServerCallContext (Metadata requestHeaders)
    {
        _requestHeaders = requestHeaders;
    }

    public static TestServerCallContext Create(Metadata? requestHeaders = null) => new (requestHeaders ?? new Metadata());

    protected override string MethodCore => "/linkforge.v1.LinkService/Resolve";
    protected override string HostCore => "localhost";
    protected override string PeerCore => "ipv4:127.0.0.1:0";
    protected override DateTime DeadlineCore => DateTime.UtcNow.AddSeconds(10);
    protected override Metadata RequestHeadersCore => _requestHeaders;
    protected override CancellationToken CancellationTokenCore => CancellationToken.None;
    protected override Metadata ResponseTrailersCore => new();
    protected override Status StatusCore {get; set;}
    protected override WriteOptions? WriteOptionsCore {get; set;}
    protected override AuthContext AuthContextCore => null!;

    protected override Task WriteResponseHeadersAsyncCore(Metadata responseHeaders) => Task.CompletedTask;
    protected override ContextPropagationToken CreatePropagationTokenCore(ContextPropagationOptions? options) => null!;
}