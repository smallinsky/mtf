CREATE TABLE happy_table (
  query VARCHAR(64) NOT NULL,
  punchline VARCHAR(64) NOT NULL,
  PRIMARY KEY (query)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO happy_table (query, punchline) VALUES ("the dirty fork", "Lucky we didn't say anything about the dirty knife");
