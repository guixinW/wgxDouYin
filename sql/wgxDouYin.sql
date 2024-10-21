CREATE TABLE `comment`  (
  `comment_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NULL,
  `deleted_at` datetime(3) NULL,
  `video_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `content` varchar(255) NOT NULL,
  PRIMARY KEY (`comment_id`)
);

CREATE TABLE `relation`  (
  `relation_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NULL,
  `deleted_at` datetime(3) NULL,
  `user_id` bigint NOT NULL,
  `to_user_id` bigint NOT NULL,
  PRIMARY KEY (`relation_id`)
);

CREATE TABLE `user`  (
  `user_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NULL,
  `deleted_at` datetime(3) NULL,
  `user_name` varchar(40) NOT NULL,
  `password` varchar(256) NOT NULL,
  `following_count` bigint NULL,
  `follower_count` bigint NULL,
  PRIMARY KEY (`user_id`)
);

CREATE TABLE `user_favorite_videos`  (
  `user_id` bigint NOT NULL,
  `video_id` bigint NOT NULL,
  PRIMARY KEY (`user_id` DESC, `video_id`)
);

CREATE TABLE `video`  (
  `video_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NULL,
  `deleted_at` datetime(3) NULL,
  `update_time` datetime(3) NULL,
  `author_id` bigint NOT NULL,
  `play_url` varchar(255) NULL,
  `cover_url` varchar(255) NULL,
  `favorite_count` bigint NULL,
  `comment_count` bigint NULL,
  `titile` varchar(50) NULL,
  PRIMARY KEY (`video_id`)
);

ALTER TABLE `comment` ADD CONSTRAINT `fk_comment_video_1` FOREIGN KEY (`video_id`) REFERENCES `video` (`video_id`);
ALTER TABLE `comment` ADD CONSTRAINT `fk_comment_user_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);
ALTER TABLE `relation` ADD CONSTRAINT `fk_relation_user_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);
ALTER TABLE `relation` ADD CONSTRAINT `fk_relation_user_2` FOREIGN KEY (`to_user_id`) REFERENCES `user` (`user_id`);
ALTER TABLE `user_favorite_videos` ADD CONSTRAINT `fk_user_favorite_videos_user_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);
ALTER TABLE `user_favorite_videos` ADD CONSTRAINT `fk_user_favorite_videos_video_1` FOREIGN KEY (`video_id`) REFERENCES `video` (`video_id`);
ALTER TABLE `video` ADD CONSTRAINT `fk_video_user_1` FOREIGN KEY (`author_id`) REFERENCES `user` (`user_id`);

