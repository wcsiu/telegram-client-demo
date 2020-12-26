build-image:
	docker build -fDockerfile -ttelegram-client-demo .

up:
	docker-compose -fdocker-compose.yml up