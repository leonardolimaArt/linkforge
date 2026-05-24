namespace Application.Abstractions;

public interface IEventPublisher
{
    Task PublishAsync<T>(
        string topic,
        string key,
        T payload,
        CancellationToken cancellationToken = default
    );
}