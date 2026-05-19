namespace Domain.ShortLink;

public class ShortLink
{
    public Guid Id {get; private set;}
    public OriginalUrl OriginalUrl {get; private set;} = null!;
    public ShortCode ShortCode {get; private set;} = null!;
    public DateTimeOffset CreatedAt {get; private set;}

    private ShortLink(){}

    private ShortLink(Guid id, OriginalUrl originalUrl, ShortCode shortCode, DateTimeOffset createdAt)
    {
        Id = id;
        OriginalUrl = originalUrl;
        ShortCode = shortCode;
        CreatedAt = createdAt;
    }

    public static ShortLink Create(OriginalUrl originalUrl, ShortCode shortCode, DateTimeOffset createdAt)
    {
        return new ShortLink(Guid.NewGuid(), originalUrl, shortCode, createdAt);
    }

}