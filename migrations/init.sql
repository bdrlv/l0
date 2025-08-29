CREATE TABLE orders (
    order_uid text PRIMARY KEY,
    track_number text NOT NULL,
    entry text,
    locale text,
    internal_signature text,
    customer_id text,
    delivery_service text,
    shardkey text,
    sm_id integer,
    date_created timestamptz NOT NULL,
    oof_shard text
);

CREATE TABLE delivery (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uid text NOT NULL UNIQUE,
    --
    name text NOT NULL,
    phone text,
    zip text,
    city text, -- можно вынести в отдельный справочник
    address text NOT NULL,
    region text, -- можно вынести в отдельный справочник
    email text
);

ALTER TABLE delivery
ADD CONSTRAINT fk_delivery_order
FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
ON DELETE CASCADE;

CREATE TABLE payment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uid text NOT NULL UNIQUE,
    --
    transaction text NOT NULL,
    request_id text,
    currency text, -- можно вынести в отдельный справочник
    provider text, -- можно вынести в отдельный справочник
    amount numeric(10,2),
    payment_dt timestamptz,
    bank text, -- можно вынести в отдельный справочник
    delivery_cost numeric(10,2),
    goods_total numeric(10,2),
    custom_fee numeric(10,2)
);

ALTER TABLE payment
ADD CONSTRAINT fk_payment_order
FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
ON DELETE CASCADE;

CREATE TABLE items (
    chrt_id integer PRIMARY KEY, -- является ли id товара? представим, что да
    name text NOT NULL,
    size text,
    nm_id integer,
    brand text -- можно вынести в отдельный справочник
);

CREATE TABLE order_items (
    order_uid text NOT NULL,
    chrt_id integer NOT NULL,
    track_number text,
    price numeric(10,2) NOT NULL,
    sale integer,
    total_price numeric(10,2) NOT NULL,
    rid text,
    status integer,
    PRIMARY KEY (order_uid, chrt_id, rid) -- rid может дублироваться?
);

ALTER TABLE order_items
ADD CONSTRAINT fk_order_items_order
FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
ON DELETE CASCADE;

ALTER TABLE order_items
ADD CONSTRAINT fk_order_items_item
FOREIGN KEY (chrt_id) REFERENCES items(chrt_id)
ON DELETE RESTRICT;