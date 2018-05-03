package main

import (
	"flag"
	"log"
	"io/ioutil"
	"strings"
	"os"
	"fmt"
	"github.com/golang-collections/collections/stack"
	"errors"
)

const variableInterpolationSymbol = "•"
const contentSegmentSeparator = variableInterpolationSymbol + variableInterpolationSymbol + variableInterpolationSymbol

const defaultVariablesFileLocation = "./specification/variables.pgen"
const defaultStructureFileLocation = "./specification/structure.pgen"
const defaultContentFileLocation = "./specification/content.pgen"
const defaultOutputDirectoryLocation = "./out"

var variablesFileLocation string
var structureFileLocation string
var contentFileLocation string
var outputDirectoryLocation string

func GetFileContents(fileLocation string) (fileContents string, err error) {
	rawFileContents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return "", err
	}

	fileContents = string(rawFileContents)

	return fileContents, nil
}

func main() () {
	variablesFileLocation = *flag.String("v", defaultVariablesFileLocation, "The location of the project variables file. Default: "+defaultVariablesFileLocation)
	structureFileLocation = *flag.String("s", defaultStructureFileLocation, "The location of the project structure file. Default: "+defaultStructureFileLocation)
	contentFileLocation = *flag.String("c", defaultContentFileLocation, "The location to of the project content file. Default: "+defaultContentFileLocation)
	outputDirectoryLocation = *flag.String("o", defaultOutputDirectoryLocation, "The location to output directory for the generated project. Default: "+defaultOutputDirectoryLocation)
	flag.Parse()

	// Parse the variables file into a map[string]string of variableName -> variableValue.
	variables, err := ParseVariablesFile()
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	// Parse the structure file into both generic structure data (variables left in) and specific structure data (variables swapped out for corresponding values).
	genericStructureData, specificStructureData, err := ParseStructureFile(variables)
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	// Parse the content file into a map[string]string of fileName -> fileContents.
	fileNameToContentMap, err := ParseContentFile(variables)
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	fmt.Println(variables)
	fmt.Println("============================================================================================================================================")
	fmt.Println(genericStructureData)
	fmt.Println("============================================================================================================================================")
	fmt.Println(specificStructureData)
	fmt.Println("============================================================================================================================================")
	fmt.Println(fileNameToContentMap)

	// Generate the project based on the structure and content map.
	err = GenerateProject(genericStructureData, specificStructureData, fileNameToContentMap)
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	fmt.Printf("SUCCESS: Project created at %v.\n", outputDirectoryLocation)
}

func ParseVariablesFile() (variables map[string]string, err error) {
	// Get variables file data.
	variablesData, err := GetFileContents(variablesFileLocation)
	if err != nil {
		return nil, err
	}

	// Parse the variables data into a map[string]string of variable-name -> variable-value.
	variables, err = ParseVariablesData(variablesData)
	if err != nil {
		return nil, err
	}

	return variables, nil
}

func ParseVariablesData(variablesData string) (variables map[string]string, err error) {
	variablesDataLines := strings.Split(variablesData, "\n")
	variables = make(map[string]string)

	for _, variablesDataLine := range variablesDataLines {
		if len(variablesDataLine) == 0 {
			continue // Then this is a blank line, skip (probably last line in file).
		}
		variableItems := strings.Split(variablesDataLine, "=")
		if len(variableItems) != 2 {
			return nil, errors.New("You must provide a single variable name, and variable value, delimited by a single \"=\". The line which causes an issue is \"" + variablesDataLine + "\".")
		}
		variables[variableItems[0]] = variableItems[1]
	}

	return variables, nil
}

func ParseStructureFile(variables map[string]string) (genericStructureData string, specificStructureData string, err error) {
	// Get the generic structure data (the structure data with interpolated variable names) from the structure file.
	genericStructureData, err = GetFileContents(structureFileLocation)
	if err != nil {
		return "", "", err
	}

	// Substitute interpolated variable names with variable values to create specific structure data.
	specificStructureData = SubstituteVariablesInStructureData(genericStructureData, variables)

	return genericStructureData, specificStructureData, nil
}

func SubstituteVariablesInStructureData(genericStructureData string, variables map[string]string) (specificStructureData string) {
	genericStructureDataLines := strings.Split(genericStructureData, "\n")

	for variableName, variableValue := range variables {
		interpolatedVariableName := variableInterpolationSymbol + variableName + variableInterpolationSymbol
		for genericStructureDataLineIndex, _ := range genericStructureDataLines {
			for strings.Index(genericStructureDataLines[genericStructureDataLineIndex], interpolatedVariableName) != -1 {
				genericStructureDataLines[genericStructureDataLineIndex] = genericStructureDataLines[genericStructureDataLineIndex][:strings.Index(genericStructureDataLines[genericStructureDataLineIndex], interpolatedVariableName)] + variableValue + genericStructureDataLines[genericStructureDataLineIndex][strings.Index(genericStructureDataLines[genericStructureDataLineIndex], interpolatedVariableName)+len(interpolatedVariableName):]
			}
		}
	}

	for _, genericStructureDataLine := range genericStructureDataLines {
		specificStructureData += genericStructureDataLine + "\n"
	}
	specificStructureData = specificStructureData[:len(specificStructureData)-1] // Shave off the extra newline added to the end.

	return specificStructureData
}

func ParseContentFile(variables map[string]string) (fileNameToContentMap map[string]string, err error) {
	// Get the content data from the content file.
	contentData, err := GetFileContents(contentFileLocation)
	if err != nil {
		return nil, err
	}

	// Build a map[string]string of file name -> its content.
	fileNameToContentMap, err = BuildFileNameToContentMap(contentData)
	if err != nil {
		return nil, err
	}

	// Substitute interpolated variable names with variable values in each entry of the fileNameToContentMap.
	fileNameToContentMap = SubstituteVariablesInFileNameToContentMap(fileNameToContentMap, variables)

	return fileNameToContentMap, nil
}

func BuildFileNameToContentMap(contentData string) (fileNameToContentMap map[string]string, err error) {
	fileNameToContentMap = make(map[string]string)

	contentSegments := strings.Split(contentData, contentSegmentSeparator+"\n")
	for _, contentSegment := range contentSegments {
		if len(contentSegment) == 0 {
			continue // Then this is a blank line, skip (probably last line in file).
		}
		contentSegmentIdentifier := contentSegment[:strings.Index(contentSegment, "\n")]
		contentSegmentData := contentSegment[strings.Index(contentSegment, "\n")+1:]

		fileNameToContentMap[contentSegmentIdentifier] = contentSegmentData
	}

	return fileNameToContentMap, nil
}

func SubstituteVariablesInFileNameToContentMap(fileNameToContentMap map[string]string, variables map[string]string) (replacedFileNameToContentMap map[string]string) {
	for variableName, variableValue := range variables {
		interpolatedVariableName := variableInterpolationSymbol + variableName + variableInterpolationSymbol
		for fileName, content := range fileNameToContentMap {
			contentLines := strings.Split(content, "\n")
			for contentLineIndex, _ := range contentLines {
				for strings.Index(contentLines[contentLineIndex], interpolatedVariableName) != -1 {
					contentLines[contentLineIndex] = contentLines[contentLineIndex][:strings.Index(contentLines[contentLineIndex], interpolatedVariableName)] + variableValue + contentLines[contentLineIndex][strings.Index(contentLines[contentLineIndex], interpolatedVariableName)+len(interpolatedVariableName):]
				}
			}

			replacedContent := ""
			for _, contentLine := range contentLines {
				replacedContent += contentLine + "\n"
			}
			replacedContent = replacedContent[:len(replacedContent)-1] // Shave off the extra newLine added to the end.
			fileNameToContentMap[fileName] = replacedContent
		}
	}
	replacedFileNameToContentMap = fileNameToContentMap

	return replacedFileNameToContentMap
}

func GenerateProject(genericStructureData string, specificStructureData string, fileNameToContentMap map[string]string) (err error) {
	genericStructureDataLines := strings.Split(genericStructureData, "\n")
	specificStructureDataLines := strings.Split(specificStructureData, "\n")

	baseDirectoryName := outputDirectoryLocation + "/" + specificStructureDataLines[0][:strings.Index(specificStructureDataLines[0], "|")] + "/"
	MakeDirectory(baseDirectoryName)

	currentSpecificPathStack := stack.New()
	currentSpecificPathStack.Push(baseDirectoryName)

	genericBaseDirectoryName := genericStructureDataLines[0]

	currentGenericPathStack := stack.New()
	currentGenericPathStack.Push(genericBaseDirectoryName)

	previousTabs := 0
	previousDirectoryTabs := 0
	for specificStructureDataLineIndex, specificStructureDataLine := range specificStructureDataLines {
		if specificStructureDataLineIndex == 0 {
			continue // This is the first line which we created above.
		} else if len(specificStructureDataLine) == 0 {
			continue // This is a blank line, skip (probably last line in file).
		} else {
			currentTabs := GetTabs(specificStructureDataLine)
			if strings.Index(specificStructureDataLine, "|") != -1 {
				specificDirectoryName := specificStructureDataLine[strings.LastIndex(specificStructureDataLine, "\t")+1:strings.Index(specificStructureDataLine, "|")] + "/"
				genericDirectoryName := genericStructureDataLines[specificStructureDataLineIndex][strings.LastIndex(genericStructureDataLines[specificStructureDataLineIndex], "\t")+1:]
				if currentTabs == previousTabs+1 {
					currentSpecificPathStack.Push(specificDirectoryName)
					currentGenericPathStack.Push(genericDirectoryName)
				} else if currentTabs <= previousTabs {
					for currentTabs < previousTabs {
						previousTabs--
						currentSpecificPathStack.Pop()
						currentGenericPathStack.Pop()
					}
					currentSpecificPathStack.Push(specificDirectoryName)
					currentGenericPathStack.Push(genericDirectoryName)
				} else {
					return errors.New("You cannot create a directory that is deep than the previous directory by more than 1 level. Every directory must have a direct parent/be a direct descendant. Check your gen file.")
				}
				newDirectoryLocation := GetRelativePath(*currentSpecificPathStack)
				MakeDirectory(newDirectoryLocation)
				previousDirectoryTabs = currentTabs
			} else {
				specificFileName := specificStructureDataLine[strings.LastIndex(specificStructureDataLine, "\t")+1:]
				genericFileName := genericStructureDataLines[specificStructureDataLineIndex][strings.LastIndex(genericStructureDataLines[specificStructureDataLineIndex], "\t")+1:]
				if currentTabs < previousTabs {
					for currentTabs < previousTabs {
						previousTabs--
						currentSpecificPathStack.Pop()
						currentGenericPathStack.Pop()
						previousDirectoryTabs--
					}
				} else if currentTabs > previousDirectoryTabs+1 {
					return errors.New("You cannot create a file that is deeper than the previous directory by more than 1 level. Every file must have a direct parent/be a direct descendant. Check your gen file.")
				}
				newFileLocation := GetRelativePath(*currentSpecificPathStack) + specificFileName
				//standardizedLocation := StandardizeLocation(newFileLocation)
				MakeFile(newFileLocation, fileNameToContentMap[GetRelativePath(*currentGenericPathStack)+genericFileName])
			}
			previousTabs = currentTabs
		}
	}

	return nil
}

func GetRelativePath(pathStack stack.Stack) (relativePath string) {
	for pathStack.Peek() != nil {
		relativePath = pathStack.Pop().(string) + relativePath
	}

	return relativePath
}

func MakeDirectory(location string) () {
	fmt.Printf("Making diretory\t\t\"%v\"\n", location)
	if err := os.Mkdir(location, 0777); err != nil {
		log.Fatalf(err.Error())
	}
}

func MakeFile(location string, content string) () {
	standardizedLocation := StandardizeLocation(location)

	fmt.Printf("Making file\t\t\"%v\",\t\tsl: \"%v\"\n", location, standardizedLocation)
	if err := ioutil.WriteFile(location, []byte(content), 0777); err != nil {
		log.Fatalf(err.Error())
	}
}

func StandardizeLocation(location string) (standardizedLocation string) {
	return strings.Replace(location[len(outputDirectoryLocation)+1:], "/", "|", -1)
}

func GetTabs(str string) (numTabs int) {
	for _, v := range str {
		if string(v) == "\t" {
			numTabs++
		}
	}

	return numTabs
}
