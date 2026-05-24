namespace Application.Events;

public record LinkCreatedEvent(
    int SchemaVersion,
    Guid Id,
    string ShortCode,
    string OriginalUrl,
    DateTimeOffset CreatedAt
);