-- 000001_create_notifications_table.up.sql
CREATE TABLE 'notifications' (
    'id' BIGINT NOT NULL AUTO_INCREMENT,
    'recipient_id' BIGINT NOT NULL,
    'sender_id' BIGINT NOT NULL,
    'notificatioin_type' VARCHAR(50) NOT NULL,
    'content' TEXT NOT NULL,
    'status' ENUM('unread', 'read') NOT NULL DEFAULT 'unread',
    'target_url' VARCHAR(255) DEFAULT NULL,
    'created_at' TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ('id'),
    INDEX 'idx_recipient_id_status_created_at' ('recipient_id'.'status', 'created_at'),
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;