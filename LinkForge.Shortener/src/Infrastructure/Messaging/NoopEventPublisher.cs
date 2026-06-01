using Application.Abstractions;
using Microsoft.Extensions.Logging;

namespace Infrastructure.Messaging;

public class NoopEventPublisher : IEventPublisher
{
    private readonly ILogger<NoopEventPublisher> _logger;

    public NoopEventPublisher(ILogger<NoopEventPublisher> logger)
    {
        _logger = logger;
        _logger.LogInformation("Kafka disabled — using NoopEventPublisher (events will be dropped)");
    }

    public Task PublishAsync<T>(string key, T payload, CancellationToken cancellationToken = default)
    {
        return Task.CompletedTask;
    }
}
