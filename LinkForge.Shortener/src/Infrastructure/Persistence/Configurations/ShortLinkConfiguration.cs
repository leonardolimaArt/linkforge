using Domain.ShortLink;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Infrastructure.Persistence.Configurations;

public class ShortLinkConfiguration : IEntityTypeConfiguration<ShortLink>
{
    public void Configure(EntityTypeBuilder<ShortLink> builder)
    {
        builder.ToTable("short_links");
        builder.HasKey(x => x.Id);
        builder.Property(x => x.Id).HasColumnName("id");
        builder.Property(x => x.ShortCode).HasColumnName("short_code").HasConversion(code => code.Value, value => new ShortCode(value)).IsRequired().HasMaxLength(32);
        builder.HasIndex(x => x.ShortCode).IsUnique();
        builder.Property(x => x.OriginalUrl).HasColumnName("original_url").HasConversion(url => url.Value, value => new OriginalUrl(value)).IsRequired().HasMaxLength(2048);
        builder.Property(x => x.CreatedAt).HasColumnName("created_at").IsRequired();
    }
}