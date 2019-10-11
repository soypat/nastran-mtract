package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var constraintTypes = []string{"RBE2", "RBE3", "RBE1"}
var reElementStart = regexp.MustCompile(`[A-Z]{2,7}[\s\d]+[\+\n]`)
var reBeamStart = regexp.MustCompile(`^[A-Z]BEAM`)
var reElementType = regexp.MustCompile(`^[A-Z]{2,7}[\d]{0,2}`)
var reElemEnd = regexp.MustCompile(`PROPERTY CARDS`)
var reEOLContinue = regexp.MustCompile(`[+]{1}$`)
var reSOLContinue = regexp.MustCompile(`^[+]{1}`)
var reInteger = regexp.MustCompile(`\s+\d+[^.A-Za-z+\$]`)
var reDecimal = regexp.MustCompile(`[-]{0,1}\d{1}\.\d{4}`)
var reNonNumerical = regexp.MustCompile(`[A-Za-z\*\+\-\s,]+`)
var reNASTRANcomment = regexp.MustCompile(`^[$]`)
var reExp = regexp.MustCompile(`[\+|\-]\d{1,2}$`)

type dims struct {
	x    float64
	y    float64
	z    float64
	t    float64
	csys int // dont think ill use it (coordinate system)
}

type node struct {
	number int
	dims
}

type element struct {
	number      int
	Type        string
	nodeIndex   []int
	collector   int
	orientation [3]float64 // Almost exclusively for beams
}

func NewElement() element {
	return element{}
}

type collectors map[string]int

type writerGroup struct {
	writer *bufio.Writer
	file   *os.File
}

type constraint struct {
	number    int
	Type      string
	master    int
	slaves    []int
	dof       string
	collector int
}

func NewConstraint() constraint {
	return constraint{}
}

//const spacedInteger string = "%d\t"
const separator string = ","
const newline string = "\n"

func readMeshCollectors(nastrandir string) (collectors, error) {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return collectors{}, err
	}
	defer dataFile.Close()
	scanner := bufio.NewScanner(dataFile)
	meshcol := collectors{}
	reElemStart := regexp.MustCompile(`ELEMENT CARDS`)
	var line = 0
	for scanner.Scan() { // Saltea las lineas hasta el comienzo de los elementos

		line++
		if reElemStart.MatchString(scanner.Text()) {
			break
		}

	}
	reMeshCol := regexp.MustCompile(`\$\*\s{2}Mesh Collector:\s{1}`)
	for scanner.Scan() {
		text := scanner.Text()
		text = text + ""
		line++
		if reElemEnd.MatchString(scanner.Text()) {
			return meshcol, nil
		} // Si encontramos con el final devolvemos lo que encontramos
		if reMeshCol.MatchString(scanner.Text()) { // I first find my mesh collector here
			collectorName := scanner.Text()[20:]
			// meshcol.tags =append(meshcol.tags, scanner.Text()[20:])
			text = scanner.Text()
			text = text + ""
			currentConstraint, currentElement, linesRead, err := readNextElement(scanner)
			if err != nil {
				return meshcol, err
			}
			line += linesRead
			if currentConstraint.slaves != nil {
				meshcol[collectorName] = currentConstraint.collector
				continue
			}
			if currentElement.nodeIndex != nil && currentConstraint.slaves == nil {
				meshcol[collectorName] = currentElement.collector
			} else {
				break
			} //We done looking for collectors
		}
	}
	return meshcol, nil
}

func writeMeshCollector(nastrandir string, Acol string, numbering int) error {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return err
	}
	writers := make(map[string]writerGroup)

	defer dataFile.Close()
	scanner := bufio.NewScanner(dataFile)
	//reCollectorName := regexp.MustCompile(Acol)
	line := 0

	for scanner.Scan() {
		text := scanner.Text()
		text = text + ""
		line++
		if strings.Contains(scanner.Text(), "$*  Mesh Collector: "+Acol) {
			scanner.Scan()
			for scanner.Scan() {
				line++
				if reNASTRANcomment.MatchString(scanner.Text()) && !(strings.Contains(scanner.Text(), "$*  Mesh Collector: "+Acol)) {
					break
				}
				constrainto, elemento, linesRead, err := readNextElement(scanner)
				if err != nil {
					return err
				}
				line += linesRead
				if elemento.nodeIndex != nil { // Si es un elemento, lo escribo a su respectivo archivo
					elementTag := generateElementTag(*elemento)
					_, present := writers[elementTag]
					if !present { // Si no hay un archivo correspondiente al elemento, lo creo
						elementFile, err := os.Create(elementTag + fmt.Sprintf("-%d", elemento.collector) + ".csv")
						if err != nil {
							return err
						}
						elementWriter := bufio.NewWriter(elementFile)
						writers[elementTag] = writerGroup{
							writer: elementWriter,
							file:   elementFile,
						}
						defer writers[elementTag].Close()
					}
					// Sigo acá, Ahora si o si tengo un archivo para escribir
					err = writeElement(elemento, writers[elementTag].writer, numbering)
					if err != nil {
						return err
					}
					err = writers[elementTag].writer.Flush()
					if err != nil {
						return err
					}
					continue
				}
				if constrainto.slaves != nil { // Si es un constraint
					_, present := writers[constrainto.Type]
					if !present { // Si no hay un archivo correspondiente al elemento, lo creo
						constraintFile, err := os.Create(constrainto.Type + "-" + Acol + ".csv")
						if err != nil {
							return err
						}
						constraintWriter := bufio.NewWriter(constraintFile)
						writers[constrainto.Type] = writerGroup{
							writer: constraintWriter,
							file:   constraintFile,
						}
						defer writers[constrainto.Type].Close()
					}
					err = writeConstraint(constrainto, writers[constrainto.Type].writer)
					writers[constrainto.Type].writer.Flush()
					if err != nil {
						return err
					}
					err = writers[constrainto.Type].writer.Flush()
					if err != nil {
						return err
					}
					break // break para no saltear el prox rigid link
				}
			}

		}

	}
	//nodeFile, err := os.Create(writedir)
	return nil
}

func writeConstraint(constrainto *constraint, writer *bufio.Writer) error {
	_, err := writer.WriteString(fmt.Sprintf("%d", constrainto.number))
	if err != nil {
		return err
	}
	_, err = writer.WriteString(fmt.Sprintf(",%d", constrainto.master))
	if err != nil {
		return err
	}
	for _, v := range constrainto.slaves {
		_, err = writer.WriteString(fmt.Sprintf(",%d", v))
		if err != nil {
			return err
		}
	}
	_, err = writer.WriteString(fmt.Sprintf("\n"))
	return err
}

func writeElement(elemento *element, writer *bufio.Writer, numbering int) error {
	// numbering == 0, ADINA.    numbering == 1  NASTRAN
	_, err := writer.WriteString(fmt.Sprintf("%d", elemento.number))
	if err != nil {
		return err
	}
	var elemIndex []int
	switch numbering {
	case 0: // ADINA
		switch len(elemento.nodeIndex) {
		case 4: // T4 (Tetraedro 4 nodos)
			elemIndex = []int{1, 2, 3, 4}
		case 10: // T10
			elemIndex = []int{1, 2, 3, 4, 5, 7, 8, 6, 10, 9}
		case 8:
			elemIndex = []int{6, 2, 3, 7, 5, 1, 4, 8}
		case 20:
			elemIndex = []int{6, 2, 3, 7, 5, 1, 4, 8, 14, 10, 15, 18, 13, 12, 16, 20, 17, 9, 11, 19}
		default:
			elemIndex = irange(1, len(elemento.nodeIndex))
		}
	default:
		elemIndex = irange(1, len(elemento.nodeIndex))
	}
	for _, v := range elemIndex {
		_, err := writer.WriteString(fmt.Sprintf(",%d", elemento.nodeIndex[v-1]))
		if err != nil {
			return err
		}
	}
	if !(elemento.orientation[0] == 0 && elemento.orientation[1] == 0 && elemento.orientation[2] == 0) {
		for v := range elemento.orientation {
			_, err := writer.WriteString(fmt.Sprintf(",%e", elemento.orientation[v]))
			if err != nil {
				return err
			}
		}
	}
	// Newline
	_, err = writer.WriteString(fmt.Sprintf("\n"))
	if err != nil {
		return err
	}
	return nil
}

func readNextElement(scanner *bufio.Scanner) (*constraint, *element, int, error) {
	eleitem := NewElement()
	constraintitem := NewConstraint()
	linesScanned := 0
	var err error
	var eleType, eleLine string
	eleType = reElementType.FindString(scanner.Text()) // Tal vez estoy parado sobre comienzo de un elemento
	if reElemEnd.MatchString(scanner.Text()) {
		return &constraintitem, &eleitem, linesScanned, nil
	}
	if eleType == "" {
		for scanner.Scan() {
			//text := scanner.Text()
			//text = text+""
			linesScanned++
			eleType = reElementType.FindString(scanner.Text())
			if reElemEnd.MatchString(scanner.Text()) {
				return &constraintitem, &eleitem, linesScanned, nil
			} // Verfificacion EOF
			if eleType == "" {
				continue
			} else {
				break
			} // Si no encuentro un elemento, sigo!
		}
	}
	eleLine = scanner.Text()
	if reEOLContinue.MatchString(eleLine) {
		for scanner.Scan() {
			linesScanned++
			eleLine = reEOLContinue.ReplaceAllString(eleLine, "") + reSOLContinue.ReplaceAllString(scanner.Text(), "")
			if reEOLContinue.MatchString(eleLine) {
				continue
			} else {
				break
			} // Llegado a este punto arme eleLine!
		}
	}
	eleLine = eleLine + " " // this space is for the regex to work consistently when last node is single digit :( not too happy about it. looking for elegant solution.
	for _, v := range constraintTypes {
		if v == eleType { //if true we are dealing with a constraint
			integerStrings := reInteger.FindAllString(eleLine, -1)
			constraintitem.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[0], ""))
			if err != nil {
				return &constraintitem, &eleitem, linesScanned, err
			}
			constraintitem.Type = v
			constraintitem.master, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[1], ""))
			if err != nil {
				return &constraintitem, &eleitem, linesScanned, err
			}
			constraintitem.dof = integerStrings[2]
			constraintitem.slaves, err = Aslicetoi(integerStrings[3:])
			return &constraintitem, &eleitem, linesScanned, err
		}
	}
	if !reElementStart.MatchString(eleLine) && reBeamStart.MatchString(eleLine) { // We are dealing with a beam
		integerStrings := reInteger.FindAllString(eleLine, -1)
		//decimalStrings := reDecimal.FindAllString(eleLine, -1)
		eleitem.orientation = [3]float64{}
		for v := range eleitem.orientation {
			eleitem.orientation[v], err = parseFortranFloat(eleLine[40+v*8 : 48+v*8])
			if err != nil {
				return &constraintitem, &eleitem, linesScanned, err
			}
		}

		eleitem.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[0], ""))
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, err
		}

		eleitem.collector, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[1], ""))
		if err != nil {
			return &constraintitem, &eleitem, linesScanned, err
		}
		eleitem.Type = "CBEAM"
		eleitem.nodeIndex, err = Aslicetoi(integerStrings[2:4])

		return &constraintitem, &eleitem, linesScanned, err

	}
	integerStrings := reInteger.FindAllString(eleLine, -1)
	integerSlice, err := Aslicetoi(integerStrings)
	if err != nil {
		return &constraintitem, &eleitem, linesScanned, err
	}
	eleitem.number = integerSlice[0]
	eleitem.Type = eleType
	eleitem.collector = integerSlice[1]
	eleitem.nodeIndex = integerSlice[2:]
	return &constraintitem, &eleitem, linesScanned, err
}

func Aslicetoi(stringSlice []string) ([]int, error) {
	var intSlice []int
	for _, v := range stringSlice {
		item, err := strconv.Atoi(reNonNumerical.ReplaceAllString(v, ""))
		if err != nil {
			return nil, err
		}
		intSlice = append(intSlice, item)
	}
	return intSlice, nil
}

func parseFortranFloat(forstr string) (float64, error) {
	var thefloat float64
	forstr = strings.Trim(forstr, " ")
	expString := reExp.FindString(forstr)
	var err error
	if expString != "" {
		thefloat, err = strconv.ParseFloat(forstr[0:5]+"E"+expString, 64)
	} else {
		thefloat, err = strconv.ParseFloat(forstr, 64)
	}
	if err != nil {
		return thefloat, nil
	}
	return thefloat, nil
}

func (meshcol collectors) KeySlice() []string {
	keys := make([]string, 0, len(meshcol))
	for k := range meshcol {
		keys = append(keys, k)
	}
	return keys
}

func writeNodosCSV(writedir string, nastrandir string) error {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return err
	}
	defer dataFile.Close()
	nodeFile, err := os.Create(writedir)
	if err != nil {
		return err
	}
	line := 1
	nodeWriter := bufio.NewWriter(nodeFile)
	defer nodeFile.Sync()
	defer nodeFile.Close()
	scanner := bufio.NewScanner(dataFile)
	reGridStart := regexp.MustCompile(`GRID\*`)
	reNodeNumber := regexp.MustCompile(`(?:GRID\*\s+)([\d]+)`)
	reNonNumerical := regexp.MustCompile(`[A-Za-z\*\+\-\s,]+`)
	reLineContinueFlag := regexp.MustCompile(`\+{1}\n*$`)
	reFloat := regexp.MustCompile(`\d{1}\.\d{4,16}E{1}[\+\-]{1}\d{2}`)
	var nodeNumberString, currentText string
	var floatStrings []string //integerStrings
	var nodoLineText string

	for scanner.Scan() {
		var currentNode node
		line++
		if reGridStart.MatchString(scanner.Text()) {
			currentText = reLineContinueFlag.ReplaceAllString(scanner.Text(), "")
			nodeNumberString = reNodeNumber.FindString(currentText)
			currentNode.number, err = strconv.Atoi(reNonNumerical.ReplaceAllString(nodeNumberString, ""))
			scanner.Scan()
			line++
			currentText = currentText + scanner.Text()
			floatStrings = reFloat.FindAllString(currentText, 4)
			assignDims(&currentNode.dims, floatStrings)
		} else { // No es un nodo, sigo!
			continue
		}
		nodoLineText = fmt.Sprintf("%d%s %e%s %e%s %e%s %e %s", currentNode.number, separator, currentNode.x, separator, currentNode.y, separator, currentNode.z, separator, currentNode.t, newline)
		_, err = nodeWriter.WriteString(nodoLineText)
		err2 := nodeWriter.Flush()
		if err != nil || err2 != nil {
			return fmt.Errorf("Error escribiendo nodos. Verificar nodos.csv")
		}
	}
	return nil
}

func assignDims(dimensions *dims, floatString []string) {
	for i := range floatString {
		switch i {
		case 0:
			dimensions.x, _ = strconv.ParseFloat(floatString[i], 64)
		case 1:
			dimensions.y, _ = strconv.ParseFloat(floatString[i], 64)
		case 2:
			dimensions.z, _ = strconv.ParseFloat(floatString[i], 64)
		case 3:
			dimensions.t, _ = strconv.ParseFloat(floatString[i], 64)
		default:
			panic("Unreachable")
		}
	}
}

func generateElementTag(elemento element) string {
	Nnodos := len(elemento.nodeIndex)
	return elemento.Type + strconv.Itoa(Nnodos)
}

func (group writerGroup) Close() {
	group.writer.Flush()

	group.file.Sync()
	group.file.Close()
}

func irange(int1 int, int2 int) []int {
	slice := make([]int, 0)
	if int1 < int2 {
		for i := int1; i <= int2; i++ {
			slice = append(slice, i)
		}
	} else {
		for i := int1; i >= int2; i-- {
			slice = append(slice, i)
		}
	}
	return slice
}
