-- name: GetByShortCode :one
SELECT id, original_url, short_code, created_at
FROM short_links
WHERE short_code = $1;