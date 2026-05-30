using Grpc.Core;
using Grpc.Core.Interceptors;

namespace Api.Grpc;

public class ApiKeyInterceptor(IConfiguration config, ILogger<ApiKeyInterceptor> logger) : Interceptor{
    private const string ApiKeyHeader = "x-api-key";

    public override async Task<TResponse> UnaryServerHandler<TRequest, TResponse>(TRequest request, ServerCallContext context, UnaryServerMethod<TRequest, TResponse> continuation)
    {
        var configuredKey = config["API_KEY"];

        if (string.IsNullOrWhiteSpace(configuredKey))
        {
            logger.LogWarning("API_KEY not configured, rejecting gRPC call to {Mathod}", context.Method);
            throw new RpcException(new Status(StatusCode.Unauthenticated, "Server misconfigured"));


        }

        var receivedKey = context.RequestHeaders.GetValue(ApiKeyHeader);

        if (string.IsNullOrEmpty(receivedKey))
        {
            logger.LogWarning("Missing {Header} on gRPC call to {Method}", ApiKeyHeader, context.Method);
            throw new RpcException(new Status(StatusCode.Unauthenticated, "Missing API Key"));
        }

        if (receivedKey != configuredKey)
        {
            logger.LogWarning("Invalid API key on gRPC call to {Method}", context.Method);
            throw new RpcException(new Status(StatusCode.Unauthenticated, "Invalid API Key"));
        }

        return await continuation(request, context);
    }
}