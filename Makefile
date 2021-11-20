bootstrap:
	docker-compose -f .devops/docker/docker-compose.yaml down -vt0 --remove-orphans
	docker-compose -f .devops/docker/docker-compose.yaml up -d
	docker-compose -f .devops/docker/docker-compose.yaml logs -f

run:
	docker-compose -f .devops/docker/docker-compose.yaml build api
	docker-compose -f .devops/docker/docker-compose.yaml build worker
	docker-compose -f .devops/docker/docker-compose.yaml up -d
up:
	docker-compose -f .devops/docker/docker-compose.yaml up -d
api-sh:
	docker-compose -f .devops/docker/docker-compose.yaml exec api sh

wrk-sh:
	docker-compose -f .devops/docker/docker-compose.yaml exec app sh