SET GOOS=solaris
SET GOOS=openbsd
SET GOOS=linux

SET GOARCH=386
SET GOARCH=amd64

@REM go build -v github.com\zew\https-server -o https-server-new
go build  -o https-server-new -v

pause