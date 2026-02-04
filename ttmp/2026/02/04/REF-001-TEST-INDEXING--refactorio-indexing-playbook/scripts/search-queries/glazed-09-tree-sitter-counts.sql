-- Tree-sitter capture counts (glazed DB)
.headers on
.mode column

SELECT COUNT(*) AS capture_count FROM ts_captures;
SELECT query_name, capture_name, COUNT(*) AS count
FROM ts_captures
GROUP BY query_name, capture_name
ORDER BY count DESC
LIMIT 10;
