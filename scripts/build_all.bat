@echo off

ECHO ======== Linux builds ========
<NUL SET /p=Building linux/amd64...   
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/amd64-linux
ECHO done

<NUL SET /p=Building linux/386...     
SET GOOS=linux
SET GOARCH=386
go build -o bin/386-linux
ECHO done


ECHO ======= Windows builds =======
<NUL SET /p=Building windows/amd64... 
SET GOOS=windows
SET GOARCH=amd64
go build -o bin/amd64-windows.exe
ECHO done

<NUL SET /p=Building windows/386...   
SET GOOS=windows
SET GOARCH=386
go build -o bin/386-windows.exe
ECHO done

ECHO ======== macOS builds ========
<NUL SET /p=Building macOS/amd64...   
SET GOOS=darwin
SET GOARCH=amd64
go build -o bin/amd64-macos
ECHO done

<NUL SET /p=Building macOS/arm64...   
SET GOOS=darwin
SET GOARCH=arm64
go build -o bin/arm64-macos
ECHO done