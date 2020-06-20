binname = emtract
buildflags = -ldflags="-s -w" -i
examplesfolder = examples

distr: win
	cp README.md README.txt
	# zip data files
	zip nastran-emtract ${examplesfolder}/nodos.csv ${examplesfolder}/CTETRA4-2.csv ${examplesfolder}/ejemplo1.dat ${examplesfolder}/ejemplo2.dat
	# zip matlab files
	zip nastran-emtract ${examplesfolder}/cargarNastran.m  ${examplesfolder}/preprocnodos.m ${examplesfolder}/bandplotx.m
	zip -j nastran-emtract bin/emtract.exe README.txt
	rm README.txt
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