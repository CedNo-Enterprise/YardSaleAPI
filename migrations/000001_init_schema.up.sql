-- Initial schema setup

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    password  TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sellers (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_items (
    id BIGSERIAL PRIMARY KEY,
    name TEXT,
    description TEXT,
	price FLOAT NOT NULL,
	status TEXT DEFAULT 'available'
);

CREATE TABLE addresses (
    id BIGSERIAL PRIMARY KEY,
    line1 TEXT NOT NULL,
    line2 TEXT,
	city TEXT NOT NULL,
	state TEXT NOT NULL,
	postalCode TEXT NOT NULL,
	country TEXT NOT NULL,
	latitude FLOAT NOT NULL,
	longitude FLOAT NOT NULL
);

CREATE TABLE saved_addresses (
    id SERIAL PRIMARY KEY,
    seller_id UUID NOT NULL,
    label TEXT,
    address_id BIGINT NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_seller FOREIGN KEY (seller_id) REFERENCES sellers(id),
    CONSTRAINT fk_address FOREIGN KEY (address_id) REFERENCES addresses(id)
);

CREATE TABLE sales (
    id UUID PRIMARY KEY,
    seller_id UUID NOT NULL,
    name TEXT,
    address_id BIGINT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'scheduled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_seller FOREIGN KEY (seller_id) REFERENCES sellers(id),
    CONSTRAINT fk_address FOREIGN KEY (address_id) REFERENCES addresses(id)
);

CREATE TABLE sale_items (
    id BIGSERIAL PRIMARY KEY,
    inventory_item_id BIGINT NOT NULL,
    name TEXT,
    price FLOAT NOT NULL,
    status TEXT DEFAULT 'available',
    sale_id UUID NOT NULL,
    CONSTRAINT fk_inventory_item FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id),
    CONSTRAINT fk_sale FOREIGN KEY (sale_id) REFERENCES sales(id)
);