all: run

run:
	./webhook-create-signed-cert.sh
	./export_cert.sh
	kubectl apply -f debug.yaml
	go run .
