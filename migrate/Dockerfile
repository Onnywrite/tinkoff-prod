FROM migrate/migrate:v4.17.1

ENTRYPOINT [ "" ]
CMD ["sh", "-c", "migrate -path /migrations -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable up"]