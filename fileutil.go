package nastran_mtract

import (
	"os"
	"path/filepath"
	"regexp"
)

func fileListCurrentDirectory(permittedStringLength int) ([]string, []string, error) {
	var files []string
	root, err := filepath.Abs("./")
	if err != nil {
		return nil, nil, err
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	// Ahora lo que hago es excluir la parte reduntante del dir
	// C:/Go/mydir/foo/myfile.exe  ----> se convierte a ---> foo/myfile.exe
	//const numberOfFiles
	var fileLength int
	maxFileLength := 0
	minFileLength := 2047

	i := 0
	var shortFileNames, actualFileNames []string
	//shortFileNames = append(files[:0:0], files...)
	//actualFileNames = append(files[:0:0], files...)

	for _, file := range files {
		fileLength = len(file)
		if fileLength > maxFileLength {
			maxFileLength = fileLength
		}
		if fileLength < minFileLength {
			minFileLength = fileLength
		}
		i++
	}
	//permittedStringLength := 54
	//i = 0
	reDir := regexp.MustCompile(`[\\]{1}[\w]{0,99}$`)
	reExecutable := regexp.MustCompile(`\.exe$`) // TODO No hace falta estos checks si voy por el otro camino
	reReadable := regexp.MustCompile(`\.txt$|\.dat$`)
	// FOR to remove base folder entry y ejecutables
	//i=0
	for i = 0; i < len(files); {
		currentString := files[i]
		if reDir.MatchString(currentString) || reExecutable.MatchString(currentString) {
			files = removeS(files, i)
			continue
		}
		if !reReadable.MatchString(currentString) {
			files = removeS(files, i)
			continue
		}
		i++
	}

	for _, file := range files {
		//if len(file) <= minFileLength {
		//	files = removeS(files, i)
		//	shortFileNames = removeS(shortFileNames, i)
		//	actualFileNames = removeS(actualFileNames, i)
		//	continue
		//}
		if len(file) > permittedStringLength+minFileLength {

			shortFileNames = append(shortFileNames, `~\…`+file[len(file)-permittedStringLength:])

		} else {
			shortFileNames = append(shortFileNames, "~"+file[minFileLength:])

		}
		actualFileNames = append(actualFileNames, "~"+file[minFileLength:])
		i++
	}
	return shortFileNames, actualFileNames, nil
}

func removeS(s []string, i int) []string {
	s = append(s[:i], s[i+1:]...)
	return s
}