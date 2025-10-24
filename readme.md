# Проект "kepr news"

## Описание проекта
веб приложение новостного форума
### Backend
- **Язык**: Golang (Echo framework)
- **Маршрутизация и агрегация запросов**: API Gateway
- **База данных**: PostgreSQL
- **Кэширование**: Redis
- **Аутентификация**: JWT tokens

### Frontend
- **Технологии**: HTML + CSS + JavaScript
- **Архитектура**: Monolit

### Инфраструктура
- **Контейнеризация**: Docker, Docker-compose
- **Мониторинг**: Prometheus + Grafana
- **Логирование**: Centralized logging

## Установка
```bash
git clone https://github.com/kempedron/keprNews
cd keprNews
sudo chmod +x start.sh
sudo ./start