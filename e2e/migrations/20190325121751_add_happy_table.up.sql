CREATE TABLE `happy_table` (
  `question` VARCHAR(64) NOT NULL,
  `punchline` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`question`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `happy_table` VALUES ("foo", "bar");
