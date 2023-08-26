run:
	docker compose -f docker-compose.yml up

test:
	docker compose -f docker-compose-testing.yml up --force-recreate -V --abort-on-container-exit

clear:
	docker compose -f docker-compose.yml down --volumes

swagger:
	swagger-codegen generate -i swagger.yml -l go-server -o .
