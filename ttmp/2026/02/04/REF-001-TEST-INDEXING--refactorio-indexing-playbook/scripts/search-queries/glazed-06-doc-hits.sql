-- Doc hits term distribution and sample rows (glazed DB)
.headers on
.mode column

SELECT term, COUNT(*) AS hits
FROM doc_hits
GROUP BY term
ORDER BY hits DESC;

SELECT f.path AS file_path, d.line, d.col, d.term, d.match_text
FROM doc_hits d
JOIN files f ON f.id = d.file_id
ORDER BY d.id DESC
LIMIT 5;
