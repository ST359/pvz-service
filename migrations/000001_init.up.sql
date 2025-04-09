CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) CHECK(role IN('employee','moderator')) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    city TEXT CHECK(city IN('Москва', 'Казань', 'Санкт-Петербург')) NOT NULL
);

CREATE TABLE IF NOT EXISTS receipts (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pvz_id UUID REFERENCES pvz(id) ON DELETE CASCADE,
    status VARCHAR(20) CHECK (status IN ('in_progress', 'close')) DEFAULT 'in_progress'
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reception_id UUID REFERENCES receipts(id) ON DELETE CASCADE,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    type TEXT CHECK (type IN ('электроника', 'одежда', 'обувь')) NOT NULL
);