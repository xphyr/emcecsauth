Simple client to connect to the EMC ECS and retrieve S3 Key

checkout the code
run go get
run go build

To build for different platforms you can do the following

For Windows:
GOOS=windows GOARCH=386 go build -o emcecsauth.exe ecsauth.go

For OSX:
GOOS=darwin GOARCH=386 go build -o emcecsauth.osx ecsauth.go
