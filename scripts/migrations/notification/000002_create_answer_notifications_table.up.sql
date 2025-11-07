-- 000002_create_answer_notifications_table.up.sql
CREATE TABLE 'answer_notifications' (
    'id' BIGINT NOT NULL AUTO_INCREMENT,
    'notification_id' BIGINT NOT NULL,
    'question_id' BIGINT NOT NULL,
    'answer_id' BIGINT NOT NULL,
    'answer_summary' VARCHAR(255) NOT NULL,
    PRIMARY KEY ('id'),
    UNIQUE KEY 'uniq_notification_id' ('notification_id'),
    FOREIGN KEY ('notification_id') REFERENCES 'notifications' ('id') ON DELETE CASCADE,
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;