#!/bin/bash

echo "🚀 Запуск пересборки Docker-проекта..."

docker-compose down
docker-compose build --no-cache
docker-compose up

echo "✅ Пересборка завершена!"