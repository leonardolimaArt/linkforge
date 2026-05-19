using Application.Abstractions;
using Domain.ShortLink;
using Microsoft.EntityFrameworkCore;

namespace Infrastructure.Persistence.Repositories;

public class ShortLinkRepository : IShortLinkRepository
{
    private readonly AppDbContext _context;
    
    public ShortLinkRepository(AppDbContext context)
    {
        _context = context;
    }


    public async Task AddAsync(ShortLink shortLink, CancellationToken cancellationToken = default)
    {
        await _context.ShortLinks.AddAsync(shortLink, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task<ShortLink?> GetByShortCodeAsync(string shortCode, CancellationToken cancellationToken = default)
    {
        return await _context.ShortLinks.FirstOrDefaultAsync(x => x.ShortCode == new ShortCode(shortCode), cancellationToken);
    }
}