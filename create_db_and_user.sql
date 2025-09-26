-- Создаем базу данных (в PostgreSQL нет IF NOT EXISTS для DATABASE)
CREATE DATABASE newDataBase;

-- Создаем пользователя, если он не существует
DO $$
BEGIN
    CREATE ROLE dbUser LOGIN PASSWORD '123456';
EXCEPTION WHEN duplicate_object THEN
    RAISE NOTICE 'Пользователь dbUser уже существует.';
END $$;

-- Предоставляем полный доступ пользователю на новую базу данных
GRANT ALL PRIVILEGES ON DATABASE newDataBase TO dbUser;