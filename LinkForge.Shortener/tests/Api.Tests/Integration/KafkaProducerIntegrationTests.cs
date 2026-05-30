using System.Text.Json;
using Application.Abstractions;
using Application.Events;
using Confluent.Kafka;
using FluentAssertions;
using Infrastructure.Messaging;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using Testcontainers.Kafka;

namespace Api.Tests.Integration;

public class KafkaProducerIntegrationTests : IAsyncLifetime
{
    private readonly KafkaContainer _kafka;
    private IEventPublisher? _publisher;

    private string _bootstrap = string.Empty;

    public KafkaProducerIntegrationTests()
    {
        _kafka = new KafkaBuilder("confluentinc/cp-kafka:7.5.0")
            .Build();
    }

    public async Task InitializeAsync()
    {
        await _kafka.StartAsync();
        _bootstrap = _kafka.GetBootstrapAddress();
        var settings = Options.Create(new KafkaSettings
        {
            BootstrapServers = _bootstrap,
            LinksCreatedTopic = "linkforge.links.created"
        });

        _publisher = new KafkaEventPublisher(settings, NullLogger<KafkaEventPublisher>.Instance);
    }

    public async Task DisposeAsync()
    {
        if (_publisher is IDisposable d) d.Dispose();
        await _kafka.DisposeAsync();
    }

    [Fact]
    public async Task Publish_Should_Send_LinkCreatedEvent_With_Correct_Payload()
    {
        var evt = new LinkCreatedEvent(
            SchemaVersion: 1,
            Id: Guid.NewGuid(),
            ShortCode: "abc12345",
            OriginalUrl: "https://google.com/integration",
            CreatedAt: DateTimeOffset.UtcNow);

        await _publisher!.PublishAsync(key: evt.ShortCode, payload: evt);
        var consumerConfig = new ConsumerConfig
        {
            BootstrapServers = _bootstrap,
            GroupId = "test-validator-" + Guid.NewGuid(),
            AutoOffsetReset = AutoOffsetReset.Earliest,
            EnableAutoCommit = false
        };

        using var consumer = new ConsumerBuilder<string, string>(consumerConfig).Build();
        consumer.Subscribe("linkforge.links.created");

        var result = consumer.Consume(TimeSpan.FromSeconds(15));
        result.Should().NotBeNull();
        result!.Message.Key.Should().Be("abc12345");

        var deserialized = JsonSerializer.Deserialize<LinkCreatedEvent>(
            result.Message.Value,
            new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower });

        deserialized.Should().NotBeNull();
        deserialized!.Id.Should().Be(evt.Id);
        deserialized.ShortCode.Should().Be(evt.ShortCode);
        deserialized.OriginalUrl.Should().Be(evt.OriginalUrl);

        consumer.Close();
    }

    [Fact]
    public async Task Publish_Should_Not_Throw_When_Topic_Auto_Created()
    {
        var evt = new LinkCreatedEvent(
            SchemaVersion: 1,
            Id: Guid.NewGuid(),
            ShortCode: "newtopic",
            OriginalUrl: "https://google.com",
            CreatedAt: DateTimeOffset.UtcNow);
        var act = async () => await _publisher!.PublishAsync(key: evt.ShortCode, payload: evt);

        await act.Should().NotThrowAsync();
    }
}
