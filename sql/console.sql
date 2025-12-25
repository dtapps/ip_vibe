-- 版本
SELECT sqlite_version();

-- 查看每个 表 的大小
SELECT
    name AS table_name,
    SUM(pgsize) AS size_bytes,
    CASE
        WHEN SUM(pgsize) >= 1073741824 THEN printf('%.2f GB', SUM(pgsize) / 1073741824.0)
        WHEN SUM(pgsize) >= 1048576     THEN printf('%.2f MB', SUM(pgsize) / 1048576.0)
        WHEN SUM(pgsize) >= 1024        THEN printf('%.2f KB', SUM(pgsize) / 1024.0)
        ELSE printf('%d B', SUM(pgsize))
    END AS size_human
FROM dbstat
WHERE name IN (SELECT name FROM sqlite_master WHERE type = 'table')
GROUP BY name
ORDER BY size_bytes DESC;

-- 查看每个 索引 的大小
SELECT
    name AS index_name,
    SUM(pgsize) AS size_bytes,
    CASE
        WHEN SUM(pgsize) >= 1073741824 THEN printf('%.2f GB', SUM(pgsize) / 1073741824.0)
        WHEN SUM(pgsize) >= 1048576     THEN printf('%.2f MB', SUM(pgsize) / 1048576.0)
        WHEN SUM(pgsize) >= 1024        THEN printf('%.2f KB', SUM(pgsize) / 1024.0)
        ELSE printf('%d B', SUM(pgsize))
    END AS size_human
FROM dbstat
WHERE name IN (SELECT name FROM sqlite_master WHERE type = 'index')
GROUP BY name
ORDER BY size_bytes DESC;

--   同时列出 表和索引 的大小
WITH table_sizes AS (
    SELECT
        name AS obj_name,
        'table' AS type,
        SUM(pgsize) AS size_bytes
    FROM dbstat
    WHERE name IN (SELECT name FROM sqlite_master WHERE type = 'table')
    GROUP BY name
),
index_sizes AS (
    SELECT
        name AS obj_name,
        'index' AS type,
        SUM(pgsize) AS size_bytes
    FROM dbstat
    WHERE name IN (SELECT name FROM sqlite_master WHERE type = 'index')
    GROUP BY name
),
all_sizes AS (
    SELECT * FROM table_sizes
    UNION ALL
    SELECT * FROM index_sizes
)
SELECT
    obj_name,
    type,
    size_bytes,
    CASE
        WHEN size_bytes >= 1073741824 THEN printf('%.2f GB', size_bytes / 1073741824.0)
        WHEN size_bytes >= 1048576     THEN printf('%.2f MB', size_bytes / 1048576.0)
        WHEN size_bytes >= 1024        THEN printf('%.2f KB', size_bytes / 1024.0)
        ELSE printf('%d B', size_bytes)
    END AS size_human
FROM all_sizes
ORDER BY size_bytes DESC;