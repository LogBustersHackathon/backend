export GOARCH=amd64
export GOOS=windows
go build -ldflags="-s -w" -o LogBustersBackendWindows.exe

export GOARCH=amd64
export GOOS=linux
go build -ldflags="-s -w" -o LogBustersBackendLinux

export GOARCH=amd64
export GOOS=darwin
go build -ldflags="-s -w" -o LogBustersBackendMacOS
