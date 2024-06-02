CREATE TYPE order_status AS ENUM ('NEW', 'REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE orders (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_number VARCHAR UNIQUE NOT NULL,
    status order_status NOT NULL,
    accrual NUMERIC(11, 2) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);