namespace Domain.ShortLink;

public sealed record ShortCode
{
    public string? Value {get;}

    public ShortCode(string value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            throw new ArgumentException("O código não pode ser vazio", nameof(value));
        }

        if(value.Length is < 4 or > 32)
        {
            throw new ArgumentException("O código deve ter tamanho entre 4 e 32 caracteres", nameof(value));
        }

        Value = value;
    }
}