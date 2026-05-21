namespace Api.Middleware;

public class ApiKeyMiddleWare(RequestDelegate next, ILogger<ApiKeyMiddleWare> logger)
{
    private const string ApiKeyHeader = "X-Api-Key";

    public async Task InvokeAsync(HttpContext context, IConfiguration config)
    {
        if(context.Request.Method == "OPTIONS") {
            await next(context);
            return;
        }

        if (context.Request.Path.StartsWithSegments("/r"))
        {
            await next(context);
            return;
        }

        var configuredKey = config["API_KEY"];

        if (string.IsNullOrWhiteSpace(configuredKey))
        {
            logger.LogWarning("ApiKey not configured, rejecting {Method} {Path}",
                context.Request.Method, context.Request.Path);
            context.Response.StatusCode = 401;
            await context.Response.WriteAsync("Unauthorized");
            return;
        }

        if (!context.Request.Headers.TryGetValue(ApiKeyHeader, out var receivedKey))
        {
            logger.LogWarning("Missing {Header} on {Method} {Path}",
                ApiKeyHeader, context.Request.Method, context.Request.Path);
            context.Response.StatusCode = 401;
            await context.Response.WriteAsync("Unauthorized");
            return;
        }

        if (receivedKey != configuredKey)
        {
            logger.LogWarning("Invalid API Key on {Method} {Path}",
                context.Request.Method, context.Request.Path);
            context.Response.StatusCode = 401;
            await context.Response.WriteAsync("Unauthorized");
            return;
        }
        await next(context);
    }
}