set GOARCH=amd64
set GOOS=windows
go build -ldflags="-s -w" -o LogBustersBackendWindows.exe

set GOARCH=amd64
set GOOS=linux
go build -ldflags="-s -w" -o LogBustersBackendLinux

set GOARCH=amd64
set GOOS=darwin
go build -ldflags="-s -w" -o LogBustersBackendMacOS
