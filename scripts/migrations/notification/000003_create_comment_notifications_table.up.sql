--- 000003_create_comment_notifications_table.up.sql
CREATE TABLE 'comment_notifications' (
    'id' BIGINT NOT NULL AUTO_INCREMENT,
    'answer_id' BIGINT NOT NULL,
    'comment_id' BIGINT NOT NULL,
    'notification_id' BIGINT NOT NULL,
    'comment_summary' VARCHAR(255) NOT NULL,
    PRIMARY KEY ('id'),
    UNIQUE KEY 'uniq_notification_id' ('notification_id'),
    FOREIGN KEY ('notification_id') REFERENCES 'notifications' ('id') ON DELETE CASCADE,
    INDEX 'idx_comment_id' ('comment_id')
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;