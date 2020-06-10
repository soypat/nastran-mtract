binname = emtract
buildflags = -ldflags="-s -w" -i

win:
	GOOS=windows GOARCH=amd64 go build ${buildflags} -o bin/${binname}.exe

win32:
	GOOS=windows GOARCH=386 go build ${buildflags} -o bin/${binname}-win32.exe

linux:
	GOOS=linux GOARCH=amd64 go build ${buildflags} -o  bin/${binname}

linux32:
	GOOS=linux GOARCH=386 go build ${buildflags} -o bin/${binname}-lin32

mac:
	GOOS=darwin GOARCH=amd64 go build ${buildflags} -o bin/${binname}-mac

all: win win32 linux linux32 mac