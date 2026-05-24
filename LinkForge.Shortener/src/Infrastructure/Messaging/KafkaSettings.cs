namespace Infrastructure.Messaging;

public class KafkaSettings
{
    public string BootstrapServers {get; set;} = string.Empty;
    public string SaslMechanism {get; set;} = string.Empty;

    public string SaslUsername {get; set;} = string.Empty;
    public string SaslPassword {get; set;} = string.Empty;
    public string LinksCreatedTopic {get; set;} = string.Empty;
}