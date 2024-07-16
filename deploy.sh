docker image prune -f

docker compose down --remove-orphans

docker compose build --no-cache

docker compose up