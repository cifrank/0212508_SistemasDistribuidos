# EJEMPLO DE COMANDOS QUE SIRVEN 
### (para el main.go no he probado comandos para lo nuevo del log)

### Le puse comentarios en el codigo q medio explican las cosas más complicadas para poder acordarme bien y q no se pierda la inspiracion jajaj

## Para POST:
$ curl -X POST localhost:8080 -d '{"value":"TGV0J3MgR28gIzEK"}' -H "Content-Type: application/json"

## Para GET:
$ curl -X GET "localhost:8080?offset=0"
________
docker build -t 0212508/logger-tests:latest .
docket --rm -it -d --name test 0212508/logger-tests:latest sh 
kind create cluster --name cluster1
kind load docker image 0212508/logger-tests:latest --name cluster1
helm create mychart
helm install logger ./mychart
si se crearon nuevos hay que checar el values.yml para poner en image nuestra imagen de docker "0212508/logger-tests" y en tag:  "latest"
(se crearon todos los archivos necesarios y etc pero no logramos ver que funcione el pod para correr el server_test.go)
