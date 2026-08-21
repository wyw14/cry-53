INSERT INTO config_types(name, definition)
VALUES
('postgres', '{"description":"PostgreSQL 数据源连接","environments":["dev","test","staging","prod"]}'::jsonb),
('http_api', '{"description":"离线环境内 HTTP 数据源","environments":["dev","test","staging","prod"]}'::jsonb),
('derived', '{"description":"引用其他数据源的派生配置","environments":["dev","test","staging","prod"]}'::jsonb)
ON CONFLICT (name) DO NOTHING;

