-- +goose Up

CREATE TABLE IF NOT EXISTS products(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gtin VARCHAR(14) NOT NULL,
    name TEXT NOT NULL,
    brand TEXT,
    category TEXT,
    package_quantity NUMERIC(10, 3),
    package_unit TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_products_package_quantity CHECK (package_quantity > 0),
    CONSTRAINT uq_products_gtin UNIQUE (gtin),
    CONSTRAINT chk_products_gtin CHECK (gtin ~ '^[0-9]{14}$'),
    CONSTRAINT chk_products_package_pair CHECK((package_quantity is NULL) = (package_unit is NULL)),
    CONSTRAINT chk_products_name CHECK (btrim(name) <> '')
);

CREATE TABLE IF NOT EXISTS product_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_code TEXT NOT NULL,
    raw_payload JSONB NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_updated_at TIMESTAMPTZ,

    CONSTRAINT uq_product_sources_provider_code UNIQUE (provider, external_code),
    CONSTRAINT chk_product_sources_provider CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_product_sources_external_code CHECK (btrim(external_code) <> '')
);

CREATE INDEX IF NOT EXISTS idx_product_sources_product_id
    ON product_sources (product_id);

CREATE TABLE IF NOT EXISTS negative_lookup_cache (
    gtin VARCHAR(14) NOT NULL,
    provider TEXT NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT pk_negative_lookup_cache PRIMARY KEY (gtin, provider),
    CONSTRAINT chk_negative_lookup_cache_gtin CHECK (gtin ~ '^[0-9]{14}$'),
    CONSTRAINT chk_negative_lookup_cache_provider CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_negative_lookup_cache_expiry CHECK (expires_at > checked_at)
);

CREATE INDEX IF NOT EXISTS idx_negative_lookup_cache_expires_at
    ON negative_lookup_cache (expires_at);

-- +goose Down
DROP TABLE IF EXISTS negative_lookup_cache;
DROP TABLE IF EXISTS product_sources;
DROP TABLE IF EXISTS products;
