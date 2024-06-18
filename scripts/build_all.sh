echo ======== Linux builds ========
echo -n "Building linux/amd64...   "
GOOS=linux GOARCH=amd64 go build -o bin/amd64-linux
echo done

echo -n "Building linux/386...     "
GOOS=linux GOARCH=386 go build -o bin/386-linux
echo done


echo ======= Windows builds =======
echo -n "Building windows/amd64... "
GOOS=windows GOARCH=amd64 go build -o bin/amd64-windows.exe
echo done

echo -n "Building windows/386...   "
GOOS=windows GOARCH=386 go build -o bin/386-windows.exe
echo done

echo ======== macOS builds ========
echo -n "Building macOS/amd64...   "
GOOS=darwin GOARCH=amd64 go build -o bin/amd64-macos
echo done

echo -n "Building macOS/arm64...   "
GOOS=darwin GOARCH=arm64 go build -o bin/arm64-macos
echo done