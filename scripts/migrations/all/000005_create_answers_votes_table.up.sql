-- 000005_create_answers_votes_table.up.sql
CREATE TABLE IF NOT EXISTS `answers_votes` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `answer_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_vote` (`answer_id`, `user_id`),
    FOREIGN KEY (`answer_id`) REFERENCES `answers`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;