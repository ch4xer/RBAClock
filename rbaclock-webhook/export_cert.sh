kubectl get secret admission-webhook-certs -o json | jq -r '.data."key.pem"' | base64 -d > key.pem
kubectl get secret admission-webhook-certs -o json | jq -r '.data."cert.pem"' | base64 -d > cert.pem
