-- name: GetByShortCode :one
SELECT id, original_url, short_code, created_at
FROM short_links
WHERE short_code = $1;

-- name: UpsertShortLink :exec
INSERT INTO short_links (id, original_url, short_code, created_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (short_code) DO NOTHING;