-- +migrate Up

CREATE TABLE `books` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `title`   varchar(191) NOT NULL UNIQUE,
  `author`   varchar(191) DEFAULT null,
  `num_available` bigint DEFAULT 1,
  `num_loaned` bigint DEFAULT 0,
  `userCreated_id` bigint NOT NULL,
  `dateCreated` timestamp DEFAULT CURRENT_TIMESTAMP,
  `userUpdated_id` bigint DEFAULT null,
  `dateUpdated` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `userDeleted_id` bigint DEFAULT null,
  `dateDeleted` timestamp DEFAULT null
);

CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `name`   varchar(191) NOT NULL UNIQUE,
  `role`   varchar(191) DEFAULT null,
  `password` varchar(191) NOT NULL,
  `userUpdated_id` bigint DEFAULT null,
  `dateCreated` timestamp DEFAULT CURRENT_TIMESTAMP,
  `dateUpdated` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `userDeleted_id` bigint DEFAULT null,
  `dateDeleted` timestamp DEFAULT null
);

CREATE TABLE `reservations` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `username`  varchar(191) NOT NULL,
  `title`  varchar(191) NOT NULL,
  `dateCreated` timestamp DEFAULT CURRENT_TIMESTAMP,
  `dateUpdated` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `userCreated_id` bigint DEFAULT null,
  `userDeleted_id` bigint DEFAULT null,
  `userUpdated_id` bigint DEFAULT null
);

ALTER TABLE `books` ADD FOREIGN KEY (`userUpdated_id`) REFERENCES `users` (`id`);
ALTER TABLE `books` ADD FOREIGN KEY (`userCreated_id`) REFERENCES `users` (`id`);
ALTER TABLE `books` ADD FOREIGN KEY (`userDeleted_id`) REFERENCES `users` (`id`);

ALTER TABLE `reservations` ADD FOREIGN KEY (`userUpdated_id`) REFERENCES `users` (`id`);
ALTER TABLE `reservations` ADD FOREIGN KEY (`userCreated_id`) REFERENCES `users` (`id`);
ALTER TABLE `reservations` ADD FOREIGN KEY (`userDeleted_id`) REFERENCES `users` (`id`);

ALTER TABLE `reservations` ADD FOREIGN KEY (`title`) REFERENCES `books` (`title`);
ALTER TABLE `reservations` ADD FOREIGN KEY (`username`) REFERENCES `users` (`name`);