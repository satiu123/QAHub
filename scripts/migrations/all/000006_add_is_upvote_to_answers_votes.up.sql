ALTER TABLE `answers_votes`
ADD COLUMN  `is_upvote` boolean NOT NULL DEFAULT false 
AFTER `user_id`;