-- Usersテーブルが存在しない場合は作成
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL UNIQUE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );

-- ストアドプロシージャでユーザーを追加
DELIMITER //
CREATE PROCEDURE generate_users(IN num_users INT)
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < num_users DO
        INSERT INTO `users` (`name`, `email`) VALUES (CONCAT('user', i), CONCAT('user', i, '@example.com'));
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;

-- プロシージャを呼び出して追加のユーザーを生成（例：10000人）
CALL generate_users(10000);
