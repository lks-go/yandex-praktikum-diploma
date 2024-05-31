CREATE TABLE operations (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_id UUID NOT NULL UNIQUE,
    amount INTEGER,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);