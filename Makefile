run:
	go run main.go

swagger:
	swagger-codegen generate -i swagger.yaml -l go-server -o .
