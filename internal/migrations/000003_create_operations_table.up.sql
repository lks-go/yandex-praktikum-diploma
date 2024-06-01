CREATE TABLE operations (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_number UUID NOT NULL,
    amount INTEGER,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX order_number_idx ON operations (order_number);