CREATE TABLE operations_history (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_id UUID NOT NULL,
    sum INTEGER,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);