-- 사용자 테이블
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 클라우드 연결 테이블
CREATE TABLE IF NOT EXISTS cloud_connections (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(10) NOT NULL CHECK (provider IN ('AWS', 'GCP')),
    name VARCHAR(255) NOT NULL,
    region VARCHAR(50),
    zone VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
    credential_file BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 클라우드 가격 테이블
CREATE TABLE IF NOT EXISTS cloud_price (
    id SERIAL PRIMARY KEY,
    cloud_name TEXT,
    region_name TEXT,
    instance_type TEXT,
    vcpu_count INTEGER,
    memory_gb INTEGER,
    operating_system TEXT,
    on_demand_price NUMERIC
);

-- 인덱스 생성
CREATE INDEX IF NOT EXISTS idx_cloud_connections_user_id ON cloud_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_cloud_connections_provider ON cloud_connections(provider); 
