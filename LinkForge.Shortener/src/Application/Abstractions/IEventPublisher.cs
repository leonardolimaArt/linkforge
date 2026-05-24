namespace Application.Abstractions;

public interface IEventPublisher
{
    Task PublishAsync<T>(
        string key,
        T payload,
        CancellationToken cancellationToken = default
    );
}