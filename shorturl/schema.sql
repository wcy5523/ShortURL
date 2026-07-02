CREATE DATABASE IF NOT EXISTS shorturl DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE shorturl;

CREATE TABLE IF NOT EXISTS users (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username    VARCHAR(50)     NOT NULL,
    email       VARCHAR(100)    NOT NULL,
    password    VARCHAR(255)    NOT NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_username (username),
    UNIQUE KEY uk_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS short_urls (
    id          BIGINT UNSIGNED NOT NULL,
    short_code  VARCHAR(16)     NOT NULL,
    original_url VARCHAR(2048)  NOT NULL,
    user_id     BIGINT UNSIGNED NOT NULL,
    visit_count BIGINT          NOT NULL DEFAULT 0,
    expire_at   DATETIME        NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY idx_short_code (short_code),
    KEY idx_user_id (user_id),
    KEY idx_expire_at (expire_at),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS visit_logs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    short_code  VARCHAR(16)     NOT NULL,
    ip          VARCHAR(64)     DEFAULT NULL,
    user_agent  VARCHAR(512)    DEFAULT NULL,
    referer     VARCHAR(1024)   DEFAULT NULL,
    visited_at  DATETIME        NOT NULL,
    PRIMARY KEY (id),
    KEY idx_short_code (short_code),
    KEY idx_visited_at (visited_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;