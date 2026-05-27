using System.Text.Json;
using Application.Abstractions;
using Application.Events;
using Confluent.Kafka;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Infrastructure.Messaging;

public class KafkaEventPublisher : IEventPublisher, IDisposable
{
    private readonly IProducer<string, string> _producer;
    private readonly ILogger<KafkaEventPublisher> _logger;
    private readonly JsonSerializerOptions _jsonOptions;
    private readonly Dictionary<Type, string> _topicByType;

    public KafkaEventPublisher(IOptions<KafkaSettings> settings, ILogger<KafkaEventPublisher> logger)
    {
        _logger = logger;
        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower
        };

        _topicByType = new Dictionary<Type, string>
        {
            {typeof(LinkCreatedEvent), settings.Value.LinksCreatedTopic}
        };

        var config = new ProducerConfig
        {
            BootstrapServers = settings.Value.BootstrapServers,
            Acks = Acks.All,
            EnableIdempotence = true,
            LingerMs = 5,
            CompressionType = CompressionType.Snappy,
            ClientId = "linkforge-shortener"
        };

        if (!string.IsNullOrEmpty(settings.Value.SaslMechanism))
        {
            config.SecurityProtocol = SecurityProtocol.SaslSsl;
            config.SaslMechanism = Enum.Parse<SaslMechanism>(settings.Value.SaslMechanism);
            config.SaslUsername = settings.Value.SaslUsername;
            config.SaslPassword = settings.Value.SaslPassword;
        }

        _producer = new ProducerBuilder<string, string>(config).Build();
    }

    public async Task PublishAsync<T>(string key, T payload, CancellationToken cancellationToken = default)
    {

        if(!_topicByType.TryGetValue(typeof(T), out var topic) || string.IsNullOrEmpty(topic))
        {
            _logger.LogError("No topic configured for event type {EventType}", typeof(T).Name);
            return;
        }

        try
        {
            var json = JsonSerializer.Serialize(payload, _jsonOptions);
            var message = new Message<string, string>
            {
                Key = key,
                Value = json
            };

            var result = await _producer.ProduceAsync(topic, message, cancellationToken);

            _logger.LogInformation("Kafka event published topic={Topic} key={Key} partition={Partition} offset={Offset}", result.Topic, key, result.Partition.Value, result.Offset.Value);

        } catch (ProduceException<string, string> ex)
        {
            _logger.LogError(ex, "Failed to publish to {Topic} with key {Key}: {Reason}", topic, key, ex.Error.Reason);

        } catch (Exception ex)
        {
            _logger.LogError(ex, "Unexpected error publishing to {Topic} with key {Key}", topic, key);
        }
    }


    public void Dispose()
    {
        _producer?.Flush(TimeSpan.FromSeconds(10));
        _producer?.Dispose();
    }
}