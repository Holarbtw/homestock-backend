-- +goose Up

CREATE TABLE IF NOT EXISTS spaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_spaces_owner_user_id UNIQUE (owner_user_id)
);

    CREATE TABLE IF NOT EXISTS inventory_items (
    space_id UUID NOT NULL REFERENCES spaces(id),
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    catalog_product_id UUID,
    gtin VARCHAR(14),
    name TEXT NOT NULL,
    brand TEXT,
    category TEXT NOT NULL,
    unit TEXT NOT NULL,
    minimum_quantity NUMERIC(19, 3),
    package_quantity NUMERIC(19, 3),
    external_image_url TEXT,
    image_object_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_inventory_items PRIMARY KEY (id),
    CONSTRAINT uq_inventory_items_space_id UNIQUE (space_id, id),
    CONSTRAINT chk_inventory_items_gtin CHECK (gtin ~ '^[0-9]{14}$'),
    CONSTRAINT chk_inventory_items_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_inventory_items_category CHECK (category IN ('FOOD', 'MEDICINE', 'HOUSEHOLD')),
    CONSTRAINT chk_inventory_items_unit CHECK (unit IN ('ml', 'g', 'pcs')),
    CONSTRAINT chk_inventory_items_minimum_quantity CHECK (
        minimum_quantity >= 0 AND minimum_quantity <= 9223372036854775.807
    ),
    CONSTRAINT chk_inventory_items_package_quantity CHECK (
        package_quantity > 0 AND package_quantity <= 9223372036854775.807
    ),
    CONSTRAINT chk_inventory_items_external_image_url CHECK (btrim(external_image_url) <> ''),
    CONSTRAINT chk_inventory_items_image_object_key CHECK (btrim(image_object_key) <> ''),
    CONSTRAINT chk_inventory_items_image_source CHECK (
        external_image_url IS NULL OR image_object_key IS NULL
    )
);

CREATE INDEX IF NOT EXISTS idx_inventory_items_gtin
    ON inventory_items (space_id, gtin)
    WHERE gtin IS NOT NULL;

CREATE TABLE IF NOT EXISTS batches (
    space_id UUID NOT NULL,
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL,
    remaining_quantity NUMERIC(19, 3) NOT NULL,
    purchased_on DATE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_on DATE,

    CONSTRAINT pk_batches PRIMARY KEY (id),
    CONSTRAINT uq_batches_item_id UNIQUE (space_id, id, item_id),
    CONSTRAINT fk_batches_item FOREIGN KEY (space_id, item_id)
        REFERENCES inventory_items (space_id, id),
    CONSTRAINT chk_batches_remaining_quantity CHECK (
        remaining_quantity >= 0 AND remaining_quantity <= 9223372036854775.807
    )
);

CREATE INDEX IF NOT EXISTS idx_batches_item_id
    ON batches (space_id, item_id);

CREATE INDEX IF NOT EXISTS idx_batches_active_expiry
    ON batches (space_id, item_id, expires_on, added_at, id)
    WHERE remaining_quantity > 0;

CREATE TABLE IF NOT EXISTS stock_operations (
    space_id UUID NOT NULL,
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL,
    actor_user_id UUID NOT NULL,
    type TEXT NOT NULL,
    quantity NUMERIC(19, 3) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_stock_operations PRIMARY KEY (id),
    CONSTRAINT uq_stock_operations_item_id UNIQUE (space_id, id, item_id),
    CONSTRAINT fk_stock_operations_item FOREIGN KEY (space_id, item_id)
        REFERENCES inventory_items (space_id, id),
    CONSTRAINT chk_stock_operations_type CHECK (type IN ('REPLENISH', 'CONSUME', 'WRITE_OFF')),
    CONSTRAINT chk_stock_operations_quantity CHECK (
        quantity > 0 AND quantity <= 9223372036854775.807
    )
);

CREATE INDEX IF NOT EXISTS idx_stock_operations_item_history
    ON stock_operations (space_id, item_id, occurred_at DESC, id);

CREATE TABLE IF NOT EXISTS stock_operation_lines (
    space_id UUID NOT NULL,
    operation_id UUID NOT NULL,
    item_id UUID NOT NULL,
    batch_id UUID NOT NULL,
    quantity NUMERIC(19, 3) NOT NULL,

    CONSTRAINT pk_stock_operation_lines PRIMARY KEY (operation_id, batch_id),
    CONSTRAINT fk_stock_operation_lines_operation FOREIGN KEY (space_id, operation_id, item_id)
        REFERENCES stock_operations (space_id, id, item_id),
    CONSTRAINT fk_stock_operation_lines_batch FOREIGN KEY (space_id, batch_id, item_id)
        REFERENCES batches (space_id, id, item_id),
    CONSTRAINT chk_stock_operation_lines_quantity CHECK (
        quantity > 0 AND quantity <= 9223372036854775.807
    )
);

CREATE INDEX IF NOT EXISTS idx_stock_operation_lines_batch_id
    ON stock_operation_lines (space_id, batch_id);

CREATE TABLE IF NOT EXISTS idempotency_records (
    space_id UUID NOT NULL REFERENCES spaces(id),
    request_key TEXT NOT NULL,
    command_type TEXT NOT NULL,
    request_hash BYTEA NOT NULL,
    response_status SMALLINT,
    response_body BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_idempotency_records PRIMARY KEY (space_id, request_key),
    CONSTRAINT chk_idempotency_records_request_key CHECK (
        btrim(request_key) <> '' AND octet_length(request_key) <= 200
    ),
    CONSTRAINT chk_idempotency_records_command_type CHECK (btrim(command_type) <> ''),
    CONSTRAINT chk_idempotency_records_request_hash CHECK (octet_length(request_hash) = 32),
    CONSTRAINT chk_idempotency_records_response_pair CHECK (
        (response_status IS NULL) = (response_body IS NULL)
    ),
    CONSTRAINT chk_idempotency_records_response_status CHECK (
        response_status BETWEEN 200 AND 299
    )
);

CREATE TABLE IF NOT EXISTS guest_imports (
    target_space_id UUID PRIMARY KEY REFERENCES spaces(id),
    source_space_id UUID NOT NULL,
    snapshot_hash BYTEA NOT NULL,
    result JSONB NOT NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_guest_imports_source_space CHECK (source_space_id <> target_space_id),
    CONSTRAINT chk_guest_imports_snapshot_hash CHECK (octet_length(snapshot_hash) = 32),
    CONSTRAINT chk_guest_imports_result CHECK (jsonb_typeof(result) = 'object')
);

-- +goose Down
DROP TABLE IF EXISTS guest_imports;
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS stock_operation_lines;
DROP TABLE IF EXISTS stock_operations;
DROP TABLE IF EXISTS batches;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS spaces;
