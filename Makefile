.PHONY: start-rabbit
start-rabbit:
	podman run -d --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management

.PHONY: tail-rabbit
tail-rabbit:
	podman run --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management

.PHONY: stop-rabbit
stop-rabbit:
	podman stop rabbitmq

.PHONY: run-server
run-server:
	go run cmd/server/*.go

.PHONY: run-client
run-client:
	go run cmd/client/*.go
