.PHONY: all hubot


all: hubot

hubot:
	cd hubot/ && docker build -t hubot:latest .

hubot-run:
	# Ensure you have a rabbit container
	docker run --rm -e HUBOT_SLACK_TOKEN=${HUBOT_SLACK_TOKEN} \
	--link rabbit:rabbit --name hubot -e \
	 RABBITMQ_URL="amqp://guest:guest@rabbit/" hubot:latest
