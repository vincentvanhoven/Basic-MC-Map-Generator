ECHO ======== Linux builds ========
ECHO -n "Building linux/amd64...   "
GOOS=linux
GOARCH=amd64
go build -o bin/amd64-linux
ECHO done

ECHO -n "Building linux/386...     "
GOOS=linux
GOARCH=386
go build -o bin/386-linux
ECHO done


ECHO ======= Windows builds =======
ECHO -n "Building windows/amd64... "
GOOS=windows
GOARCH=amd64
go build -o bin/amd64-windows.exe
ECHO done

ECHO -n "Building windows/386...   "
GOOS=windows
GOARCH=386
go build -o bin/386-windows.exe
ECHO done

ECHO ======== macOS builds ========
ECHO -n "Building macOS/amd64...   "
GOOS=darwin
GOARCH=amd64
go build -o bin/amd64-macos
ECHO done

ECHO -n "Building macOS/arm64...   "
GOOS=darwin
GOARCH=arm64
go build -o bin/arm64-macos
ECHO done