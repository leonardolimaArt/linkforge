using Application.UseCases.ResolveShortLink;
using Google.Protobuf.WellKnownTypes;
using Grpc.Core;
using LinkForge.Grpc.V1;

namespace Api.Grpc;

public class LinkServiceImpl(ResolveShortLinkHandler resolveHandler, ILogger<LinkServiceImpl> logger) : LinkService.LinkServiceBase
{
    public override async Task<ResolveResponse> Resolve(ResolveRequest request, ServerCallContext context)
    {
        if (string.IsNullOrWhiteSpace(request.ShortCode))
        {
            throw new RpcException(new Status(StatusCode.InvalidArgument, "short_code is required"));
        }

        var result = await resolveHandler.HandleAsync(new ResolveShortLinkQuery(request.ShortCode), context.CancellationToken);

        if (result is null)
        {
            logger.LogInformation("Resolve NotFound for {ShortCode}", request.ShortCode);
            throw new RpcException(new Status(StatusCode.NotFound, $"short code '{request.ShortCode}' not found"));
        }

        return new ResolveResponse
        {
            Id = result.Id == Guid.Empty ? string.Empty : result.Id.ToString(),
            ShortCode = result.ShortCode,
            OriginalUrl = result.OriginalUrl,
            CreatedAt = result.CreatedAt == default ? null : Timestamp.FromDateTimeOffset(result.CreatedAt)
        };
    }
}