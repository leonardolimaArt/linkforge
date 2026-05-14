namespace Domain.ShortLink;

public sealed record OriginalUrl
{
    public string? Value {get;}

    public OriginalUrl(string value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            throw new ArgumentException("A URL original não pode ser vazia", nameof(value));
        }

        if (!Uri.TryCreate(value, UriKind.Absolute, out var uri))
        {
            throw new ArgumentException("A URL original deve ser absoluta/completa", nameof(value));
        }

        if(uri.Scheme is not "http" and not "https")
        {
            throw new ArgumentException("A URL original deve usar HTTP ou HTTPS", nameof(value));
        }

        Value = value;
    }
}