dev-build:
	cd docker && docker compose --env-file .env.dev up --build -d

dev-build-no-cache:
	cd docker && docker compose --env-file .env.dev build --no-cache

dev-up:
	cd docker && docker compose --env-file .env.dev up -d

dev-down:
	cd docker && docker compose --env-file .env.dev down

dev-stop:
	cd docker && docker stop ct-backend-dev && docker rm ct-backend-dev

dev-restart: dev-stop dev-up
	@echo "Restarting dev completed."

