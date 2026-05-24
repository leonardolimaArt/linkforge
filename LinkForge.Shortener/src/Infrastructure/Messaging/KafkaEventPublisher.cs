using System.Text.Json;
using Application.Abstractions;
using Confluent.Kafka;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Infrastructure.Messaging;

public class KafkaEventPublisher : IEventPublisher, IDisposable
{
    private readonly IProducer<string, string> _producer;
    private readonly ILogger<KafkaEventPublisher> _logger;
    private readonly JsonSerializerOptions _jsonOptions;

    public KafkaEventPublisher(IOptions<KafkaSettings> settings, ILogger<KafkaEventPublisher> logger)
    {
        _logger = logger;
        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower
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

    public async Task PublishAsync<T>(string topic, string key, T payload, CancellationToken cancellationToken = default)
    {
        try
        {
            var json = JsonSerializer.Serialize(payload, _jsonOptions);
            var message = new Message<string, string>
            {
                Key = key,
                Value = json
            };

            var result = await _producer.ProduceAsync(topic, message, cancellationToken);

            _logger.LogDebug("Published to {Topic} partition {Partition} offset {Offset}", result.Topic, result.Partition.Value, result.Offset.Value);

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