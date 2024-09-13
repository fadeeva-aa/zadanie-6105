CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
);


CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
);


CREATE TYPE author_type AS ENUM (
    'Organization',
    'User'
);


CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
	'Canceled',
	'Approved',
	'Rejected'
);


CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE IF NOT EXISTS tender (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
	type service_type,
	status tender_status,
	organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
	version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS tender_archive (
	id UUID NOT NULL,
	name VARCHAR(100) NOT NULL,
    description TEXT,
	type service_type,
	status tender_status,
	organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
	version INTEGER NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
	PRIMARY KEY(id, version)
);


CREATE OR REPLACE FUNCTION archive_tender()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO tender_archive (id, name, description, type, status, organization_id, version, created_at, updated_at)
    VALUES (new.id, new.name, new.description, new.type, new.status, new.organization_id, new.version, new.created_at, new.updated_at);
    RETURN new;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER trg_archive_tender
AFTER INSERT OR UPDATE ON tender
FOR EACH ROW
EXECUTE PROCEDURE archive_tender();


CREATE TABLE IF NOT EXISTS bid (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
	status bid_status,
	tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
	author_type author_type,
	author_id UUID NOT NULL,
	version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS bid_archive (
	id UUID NOT NULL,
	name VARCHAR(100) NOT NULL,
    description TEXT,
	status bid_status,
	tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
	author_type author_type,
	author_id UUID NOT NULL,
	version INTEGER NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
	PRIMARY KEY(id, version)
);


CREATE OR REPLACE FUNCTION archive_bid()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO bid_archive (id, name, description, status, tender_id, author_type, author_id, version, created_at, updated_at)
    VALUES (new.id, new.name, new.description, new.status, new.tender_id, new.author_type, new.author_id, new.version, new.created_at, new.updated_at);
    RETURN new;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER trg_archive_bid
AFTER INSERT OR UPDATE ON bid
FOR EACH ROW
EXECUTE PROCEDURE archive_bid();


CREATE TABLE IF NOT EXISTS bid_feedback (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	bid_id UUID REFERENCES bid(id) ON DELETE CASCADE,
	user_id UUID REFERENCES employee(id) ON DELETE CASCADE,
	description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS bid_approved_decision (
	bid_id UUID REFERENCES bid(id) ON DELETE CASCADE,
	user_id UUID REFERENCES employee(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(bid_id, user_id)
);


