.PHONY: start-rabbit
start-rabbit:
	podman run -d --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management

.PHONY: stop-rabbit
stop-rabbit:
	podman stop rabbitmq
