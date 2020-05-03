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
var bcTypes = []string{"SPC", "FORCE"}
var reElementStart = regexp.MustCompile(`[A-Z]{2,7}[\s\d]+[\+\n]`)
var reSPCStart = regexp.MustCompile(`^SPC`)
var reForceStart = regexp.MustCompile(`^FORCE`)
var reBeamStart = regexp.MustCompile(`^[A-Z]BEAM`)
var reElementType = regexp.MustCompile(`^[A-Z]{2,7}[\d]{0,2}`)
var reElemEnd = regexp.MustCompile(`PROPERTY CARDS`)
var reEOLContinue = regexp.MustCompile(`[+]{1}$`)
var reSOLContinue = regexp.MustCompile(`^[+]{1}`)
var reInteger = regexp.MustCompile(`\s+\d+[^.A-Za-z+\$]`)
//var reDecimal = regexp.MustCompile(`[-]{0,1}\d{1}\.\d{4}`)
var reNonNumerical = regexp.MustCompile(`[A-Za-z\*\+\-\s,]+`)
var reNASTRANcomment = regexp.MustCompile(`^[$]`)
var reExp = regexp.MustCompile(`[\+|\-]\d{1,2}$`)
var reMeshCol = regexp.MustCompile(`\$\*\s{2}Mesh Collector:\s{1}`)
var reSPCCol = regexp.MustCompile(`\$\*\s{2}Constraint:\s{1}`)
var reLoadCol = regexp.MustCompile(`\$\*\s{2}Load:\s{1}`)
var reEOF = regexp.MustCompile(`ENDDATA`)
//ELEMENT NUMBERING
//var ADINA_N = [][]int{{1, 2, 3, 4}, {1, 2, 3, 4, 5, 7, 8, 6, 10, 9}, {6, 2, 3, 7, 5, 1, 4, 8}, {6, 2, 3, 7, 5, 1, 4, 8, 14, 10, 15, 18, 13, 12, 16, 20, 17, 9, 11, 19}}
var ADINA = map[int][]int{4:{1, 2, 3, 4},10:{1, 2, 3, 4, 5, 7, 8, 6, 10, 9},8:{6, 2, 3, 7, 5, 1, 4, 8},20:{6, 2, 3, 7, 5, 1, 4, 8, 14, 10, 15, 18, 13, 12, 16, 20, 17, 9, 11, 19}}

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
	orientation [3]float64 // Almost exclusively for beams and shells
}

func newElement() element {
	return element{}
}

type collectors map[string]entity

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
	// boundary condition for forces
	dims
	magnitude float64
}

// DEFINE ENTITY INTEFACE describes both constraints and elements
type entity interface {
	getType() string
	getNumber() int
	getCollector() int
	isElement() bool
	generateUniqueTag() string
	getConnections() []int
	getAssociatedVector() (int ,[3]float64)
	getDOF() string
}
func (e element)isElement() bool {
	return true
}
func (c constraint)isElement() bool {
	return false
}
func (c constraint)getType() string {
	return c.Type
}
func (e element)getType() string {
	return e.Type
}
func (c constraint)getNumber() int {
	return c.number
}
func (e element)getNumber() int {
	return e.number
}
func (c constraint)getCollector() int {
	return c.collector
}
func (e element)getCollector() int {
	return e.collector
}
func (e element)getConnections() []int {
	return e.nodeIndex
}
func (c constraint)getConnections() []int {
	var con []int
	con = append(con,c.master)
	if len(c.slaves)>0 {
		for _, v := range c.slaves {
			con = append(con,v)
		}
	}
	return con
}
func (e element)getAssociatedVector() (int, [3]float64) {
	return 0,e.orientation
}
func (c constraint)getAssociatedVector() (int ,[3]float64) {
	var vec [3]float64
	vec[0] = c.x * c.magnitude
	vec[1] = c.y * c.magnitude
	vec[2] = c.z * c.magnitude
	return c.csys,vec
}
func (e element)getDOF() string {
	return ""
}
func (c constraint)getDOF() string {
	return c.dof
}

func newConstraint() constraint {
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
	var finishedElements bool
	for scanner.Scan() {
		text := scanner.Text() + ""; line++
		collectorName := parseCollectorName(text) // =="" if not collector name
		var linesRead int

		if reElemEnd.MatchString(scanner.Text()) {
			finishedElements = true
		}
		if finishedElements && collectorName != "" {
			meshcol[collectorName] , linesRead, err = readNextElement(scanner)
		}
		if !finishedElements && collectorName != ""  { // I first find my mesh collector here
			meshcol[collectorName] , linesRead, err = readNextElement(scanner)
		}
		line += linesRead
		if err != nil {
			return meshcol, err
		}
	}
	return meshcol, nil
}

func writeEntity(E entity, writer *bufio.Writer, numbering int) error {
	defer writer.WriteString("\n") // Deferamos el newline del final de la tabla
	// Print DOF and entity number
	if E.getDOF() != "" { // DOF puede ser vacio
		writer.WriteString(fmt.Sprintf("%s , %d", E.getDOF(), E.getNumber()))
	} else {
		writer.WriteString(fmt.Sprintf("%d", E.getNumber()))
	}

	// Print connections
	var connections = E.getConnections()
	var connectionIndex []int
	if numbering==0 && len(ADINA[len(connections)])>0 && E.isElement() {
		connectionIndex = ADINA[len(connections)]
	} else {
		connectionIndex = irange(1,len(connections))
	}
	for _, v := range connectionIndex {
		_, err := writer.WriteString(fmt.Sprintf(",%d", connections[v-1]))
		if err != nil {
			return err
		}
	}
	// Print associated vector if present
	var vec [3]float64
	var _ int //  would be csys
	_, vec = E.getAssociatedVector()
	if !(vec[0] == 0 && vec[1] == 0 && vec[2] == 0) {
		for v := range vec {
			_, err := writer.WriteString(fmt.Sprintf(",%e", vec[v]))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeCollector(nastrandir string, Acol string, numbering int) error {
	dataFile, err := os.Open(nastrandir)
	if err != nil {
		return err
	}
	writers := make(map[string]writerGroup)

	defer dataFile.Close()
	scanner := bufio.NewScanner(dataFile)
	//reCollectorName := regexp.MustCompile(Acol)
	line := 0
	var text string
	for scanner.Scan() {
		text = scanner.Text() + ""; line++
		collectorName := parseCollectorName(text)

		if collectorName == Acol  {
			for scanner.Scan() {
				text = scanner.Text() + ""; line++
				var E entity
				collectorFound := parseCollectorName(text)
				if (collectorFound != Acol && collectorFound != "") || reEOF.MatchString(text) || reElemEnd.MatchString(text)  {
					break
				}
				if reNASTRANcomment.MatchString(text) {
					continue
				}
				E, linesRead, err := readNextElement(scanner)
				text = scanner.Text() + ""; line += linesRead
				if err != nil {
					return err
				}
				entityTag := E.generateUniqueTag()
				_, present := writers[entityTag]
				if !present { // Si no hay un archivo correspondiente a la entidad, lo creo
					elementFile, err := os.Create(fmt.Sprintf("%s.csv",entityTag) )
					if err != nil {
						return err
					}
					elementWriter := bufio.NewWriter(elementFile)
					writers[entityTag] = writerGroup{
						writer: elementWriter,
						file:   elementFile,
					}
					defer writers[entityTag].Close()
				} // Sigo acá, Ahora si o si tengo un archivo para escribir
				err = writeEntity(E, writers[entityTag].writer, numbering)
				if err != nil {
					return err
				}
				err = writers[entityTag].writer.Flush()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func readNextElement(scanner *bufio.Scanner) (entity, int, error) {
	var E entity
	eleitem := newElement()
	constraintitem := newConstraint()
	linesScanned := 0
	var err error
	var eleType, eleLine string
	eleType = reElementType.FindString(scanner.Text()) // Tal vez estoy parado sobre comienzo de un elemento
	if reElemEnd.MatchString(scanner.Text()) {
		return E, linesScanned, nil
	}
	if eleType == "" {
		for scanner.Scan() {
			if reNASTRANcomment.MatchString(scanner.Text())  {
				continue
			}
			linesScanned++
			eleType = reElementType.FindString(scanner.Text())
			if reElemEnd.MatchString(scanner.Text()) {
				return E, linesScanned, nil
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
				return E, linesScanned, err
			}
			constraintitem.Type = v
			constraintitem.master, err = strconv.Atoi(reNonNumerical.ReplaceAllString(integerStrings[1], ""))
			if err != nil {
				return E, linesScanned, err
			}
			constraintitem.dof = integerStrings[2]
			constraintitem.slaves, err = Aslicetoi(integerStrings[3:])
			E = constraintitem
			return E, linesScanned, err
		}
	}

	for _, v := range bcTypes {
		if v == eleType { //if true we are dealing with a boundary condition
			splitline := splitFortranLine(eleLine)
			constraintitem.Type = strings.TrimSpace(splitline[0])
			constraintitem.collector, err = strconv.Atoi(reNonNumerical.ReplaceAllString(splitline[1], ""))
			if err != nil {
				return E, linesScanned, err
			}
			constraintitem.master , err = strconv.Atoi(strings.TrimSpace(splitline[2]))
			if err != nil {
				return E, linesScanned, err
			}

			if eleType=="SPC" {
				constraintitem.dof  = strings.TrimSpace(splitline[3])
				constraintitem.magnitude, err = parseFortranFloat(splitline[4])
			}
			if eleType=="FORCE" {
				constraintitem.csys, err  = strconv.Atoi(reNonNumerical.ReplaceAllString(splitline[3], ""))
				if err != nil {
					return E, linesScanned, err
				}
				constraintitem.magnitude, err = parseFortranFloat(splitline[4])
				if err != nil {
					return E, linesScanned, err
				}
				constraintitem.x, err = parseFortranFloat(splitline[5])
				if err != nil {
					return E, linesScanned, err
				}
				constraintitem.y, err = parseFortranFloat(splitline[6])
				if err != nil {
					return E, linesScanned, err
				}
				constraintitem.z, err = parseFortranFloat(splitline[7])
			}
			if err != nil {
				return E, linesScanned, err
			}
			E = constraintitem
			return E, linesScanned, err
		}
	}


	if !reElementStart.MatchString(eleLine) && reBeamStart.MatchString(eleLine) { // We are dealing with a beam
		integerList := []int{} // two nodes in integerStrings

		for i:=0; i<=3 ; i++ {
			myCurrentInt, err := strconv.Atoi(reNonNumerical.ReplaceAllString(eleLine[(i+1)*8:(i+2)*8], ""))
			if err != nil {
				return E, linesScanned, err
			}
			integerList = append(integerList, myCurrentInt ) // genero lista de BEAM integers. fixed width opportunity
		}
		eleitem.number = integerList[0]
		eleitem.collector = integerList[1]
		eleitem.nodeIndex = integerList[2:4]
		eleitem.orientation = [3]float64{}
		for v := range eleitem.orientation {
			eleitem.orientation[v], err = parseFortranFloat(eleLine[40+v*8 : 48+v*8])
			if err != nil {
				return E, linesScanned, err
			}
		}
		eleitem.Type = "CBEAM"
		E = eleitem
		return E, linesScanned, err
	}
	integerStrings := reInteger.FindAllString(eleLine, -1)
	integerSlice, err := Aslicetoi(integerStrings)
	if err != nil {
		return E, linesScanned, err
	}
	eleitem.number = integerSlice[0]
	eleitem.Type = eleType
	eleitem.collector = integerSlice[1]
	eleitem.nodeIndex = integerSlice[2:]
	E = eleitem
	return E, linesScanned, err
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
	reFloat := regexp.MustCompile(`[-]{0,1}\d{1}\.\d{4,16}E{1}[\+\-]{1}\d{2}`)
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

func (constrainto constraint)generateUniqueTag() string {
	return fmt.Sprintf(  "%s-%d",constrainto.Type,  constrainto.collector)
}

func (elemento element)generateUniqueTag() string {
	Nnodos := len(elemento.nodeIndex)
	return fmt.Sprintf(  "%s%d-%d",elemento.Type, Nnodos, elemento.collector)
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

func splitFortranLine(linestr string) []string {
	var N = len(linestr)
	split := []string{}
	var current = ""
	for i:=0; i<N; i++ {
		current = current + string(linestr[i])
		if (i+1)%8 == 0  {
			split = append(split,current)
			current = ""
		}
	}
	if len(current)>0 {
		split = append(split,current)
	}
	return split
}

func parseCollectorName(text string) string {
	var collectorName string
	if reMeshCol.MatchString(text) || reSPCCol.MatchString(text) || reLoadCol.MatchString(text) {
		collectorName = text[strings.Index(text,":")+2:]
	}
	return collectorName
}
