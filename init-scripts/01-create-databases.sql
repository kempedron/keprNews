-- Создаем базу данных для auth-service
CREATE DATABASE auth_db;

-- Создаем базу данных для article-service
CREATE DATABASE article_db;

-- Предоставляем права пользователю dbUser
GRANT ALL PRIVILEGES ON DATABASE auth_db TO dbUser;
GRANT ALL PRIVILEGES ON DATABASE article_db TO dbUser;