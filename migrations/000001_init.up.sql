CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) CHECK(role IN('employee','moderator')) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pvzs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    city TEXT CHECK(city IN('Москва', 'Казань', 'Санкт-Петербург')) NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    pvz_id UUID REFERENCES pvzs(id) ON DELETE CASCADE,
    status VARCHAR(20) CHECK (status IN ('in_progress', 'close')) DEFAULT 'in_progress'
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reception_id UUID REFERENCES receptions(id) ON DELETE CASCADE,
    date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    type TEXT CHECK (type IN ('электроника', 'одежда', 'обувь')) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_receptions_date ON receptions (date);
CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions (pvz_id);
CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products (reception_id);

CREATE OR REPLACE FUNCTION get_pvz_with_receptions_paginated(
    start_date TIMESTAMP DEFAULT NULL,
    end_date TIMESTAMP DEFAULT NULL,
    page_limit INT DEFAULT 10,
    page_offset INT DEFAULT 0
) RETURNS TABLE (
    pvz_id UUID,
    city TEXT,
    registration_date TIMESTAMPTZ,
    receptions JSON
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered_pvzs AS (
        SELECT p.id, p.city, p.registration_date
        FROM pvzs p
        ORDER BY p.registration_date DESC
        LIMIT page_limit
        OFFSET page_offset
    )
    SELECT 
        p.id AS pvz_id,
        p.city,
        p.registration_date,
        CASE 
            WHEN COUNT(r.id) = 0 THEN NULL
            ELSE (
                SELECT json_agg(
                    json_build_object(
                        'reception', json_build_object(
                            'dateTime', r.date,
                            'id', r.id,
                            'pvzId', r.pvz_id,
                            'status', r.status
                        ),
                        'products', (
                            SELECT COALESCE(
                                json_agg(
                                    json_build_object(
                                        'dateTime', pr.date,
                                        'id', pr.id,
                                        'receptionId', pr.reception_id,
                                        'type', pr.type
                                    )
                                ),
                                '[]'::json
                            )
                            FROM products pr
                            WHERE pr.reception_id = r.id
                        )
                    )
                )
                FROM receptions r
                WHERE r.pvz_id = p.id
                AND (start_date IS NULL OR r.date >= start_date)
                AND (end_date IS NULL OR r.date <= end_date)
            )
        END AS receptions
    FROM filtered_pvzs p
    LEFT JOIN receptions r ON r.pvz_id = p.id
        AND (start_date IS NULL OR r.date >= start_date)
        AND (end_date IS NULL OR r.date <= end_date)
    GROUP BY p.id, p.city, p.registration_date;
END;
$$ LANGUAGE plpgsql;