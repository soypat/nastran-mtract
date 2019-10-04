package nastran_mtract



func main() {
	fileListWidth := 10

	displayedFileNames, _, err := fileListCurrentDirectory(fileListWidth - 6)
	if err != nil {
		panic("Implement error message for failing to read files!")
	}


	fileList := NewMenu()
	myPoller := NewPoller()
	myPoller.menu = &fileList
	fileList.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{1, 2, 0})
	fileList.options = displayedFileNames
	fileList.title = "Archivos disponibles"
	InitMenu(&fileList)

}
