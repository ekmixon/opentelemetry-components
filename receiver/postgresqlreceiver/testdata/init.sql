CREATE USER otel WITH PASSWORD 'otel';
GRANT SELECT ON pg_stat_database TO otel;

CREATE TABLE table1 ();
CREATE TABLE table2 ();