CREATE TABLE IF NOT EXISTS subscription
(
    chat_id    INT PRIMARY KEY NOT NULL,
    created_at BIGINT          NOT NULL,
    period     INT             NOT NULL
)