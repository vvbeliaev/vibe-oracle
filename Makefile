b:
	docker compose -f compose.local.yml build

up:
	docker compose -f compose.local.yml up -d

dev:
	make b
	make up

meili:
	docker compose -f compose.local.yml up meilisearch -d