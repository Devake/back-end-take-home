Build steps:
1, install go
2, run "go build -o back-end-take-home main.go" in the project directory
	
Usage:
back-end-take-home <data file directory> <server port>

Example:
back-end-take-home ./data/full 8409

Endpoint URL:
http://back-end-take-home.hancco.net.:8409/backendTest?origin=<origin 3 digit code>&destination=<destination 3 digit code>
Running on a AWS EC2 t2.micro instance with Debian 9

To do:
- load data into SQL database
- save metadata in caching server
- move data types to another file

Troubleshooting:
To build for linux on a window machine:
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLE=0
go build -o back-end-take-home main.go