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

const defaultGenerationFileLocation = "./assets/project.gen"
const defaultContentFilesLocation = "./content-files"
const defaultProjectOutputLocation = "./project-output"

func main() () {
	var err error

	genFileLocation := *flag.String("generation-file-location", defaultGenerationFileLocation, "The location of the project generation specification file. Default: "+defaultGenerationFileLocation)
	contentFilesLocation := *flag.String("content-files-location", defaultContentFilesLocation, "The location to of the content files for the generated project structure. Default: "+defaultContentFilesLocation)
	outputLocation := *flag.String("project-output-location", defaultProjectOutputLocation, "The location to put the generated file structure. Default: "+defaultProjectOutputLocation)
	flag.Parse()

	rawGenData, err := ioutil.ReadFile(genFileLocation)
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}
	genData := string(rawGenData)

	variables, err := ParseVariabless(genData)
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	genData, err = ReplaceVariables(genData, variables)
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	err = BuildProject(genData, contentFilesLocation, outputLocation)
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	fmt.Printf("%v\n", "Project created.")
}

func ParseVariabless(genData string) (variables map[string]string, err error) {
	genDataLines := strings.Split(genData, "\n")
	variables = make(map[string]string, 0)

	i := 0
	for genDataLines[i] != "" && i < len(genDataLines) {
		lineItems := strings.Split(genDataLines[i], "=")
		variables[lineItems[0]] = lineItems[1]

		i++
	}

	return variables, nil
}

func ReplaceVariables(genData string, variables map[string]string) (replacedGenData string, err error) {
	genDataLines := strings.Split(genData, "\n")
	genDataLines = genDataLines[len(variables)+1:]

	for variableName, variableValue := range variables {
		for genDataLineIndex, genDataLine := range genDataLines {
			if strings.Index(genDataLine, "~"+variableName+"~") != -1 {
				genDataLines[genDataLineIndex] = genDataLine[0:strings.Index(genDataLine, "~"+variableName+"~")] + variableValue + genDataLine[strings.Index(genDataLine, "~"+variableName+"~")+len(variableName)+2:]
			}
		}
	}

	for _, genDataLine := range genDataLines {
		replacedGenData += genDataLine + "\n"
	}
	replacedGenData = replacedGenData[:len(replacedGenData)-1]

	return replacedGenData, nil
}

func BuildProject(genData string, contentFilesLocation string, outputLocation string) (err error) {
	genDataLines := strings.Split(genData, "\n")

	baseDirectoryName := outputLocation + "/" + genDataLines[0][:strings.Index(genDataLines[0], "/")+1]
	MakeDirectory(baseDirectoryName)

	currentPathStack := stack.New()
	currentPathStack.Push(baseDirectoryName)

	previousTabs := 0
	previousDirectoryTabs := 0
	for genDataLineIndex, genDataLine := range genDataLines {
		if genDataLineIndex == 0 {
			continue
		} else {
			currentTabs := GetTabs(genDataLine)
			fmt.Printf("%v %v\n", previousTabs, currentTabs)
			if strings.Index(genDataLine, "/") != -1 {
				directoryName := genDataLine[strings.LastIndex(genDataLine, "\t")+1 : strings.Index(genDataLine, "/")+1]
				if currentTabs == previousTabs+1 {
					currentPathStack.Push(directoryName)
				} else if currentTabs <= previousTabs {
					for currentTabs < previousTabs {
						previousTabs--
						currentPathStack.Pop()
					}
					currentPathStack.Push(directoryName)
				} else {
					return errors.New("You cannot create a directory that is deep than the previous directory by more than 1 level. Every directory must have a direct parent/be a direct descendant. Check your gen file.")
				}
				newDirectoryLocation := GetRelativePath(*currentPathStack)
				MakeDirectory(newDirectoryLocation)
				previousDirectoryTabs = currentTabs
			} else {
				fileName := genDataLine[strings.LastIndex(genDataLine, "\t")+1:]
				if currentTabs < previousTabs {
					for currentTabs < previousTabs {
						previousTabs--
						currentPathStack.Pop()
						previousDirectoryTabs--
					}
				} else if currentTabs > previousDirectoryTabs+1 {
					return errors.New("You cannot create a file that is deeper than the previous directory by more than 1 level. Every file must have a direct parent/be a direct descendant. Check your gen file.")
				}
				newFileLocation := GetRelativePath(*currentPathStack) + fileName
				MakeFile(newFileLocation)
				contentFileLocation := defaultContentFilesLocation + "/" + StandardizeLocation(newFileLocation) + ".content"
				fmt.Printf("%v\n", contentFileLocation)
				// TODO: Variable substitute. Do this by extracting the data, not by replacing it in the file.
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

func MakeFile(location string) () {
	fmt.Printf("Making file\t\t\"%v\"\n", location)
	if err := ioutil.WriteFile(location, []byte(""), 0777); err != nil {
		log.Fatalf(err.Error())
	}
}

func StandardizeLocation(location string) (standardizedLocation string) {
	standardizedLocation = strings.Replace(location[2:], "/", "_", -1)
	return standardizedLocation
}

func GetTabs(str string) (numTabs int) {
	for _, v := range str {
		if string(v) == "\t" {
			numTabs++
		}
	}

	return numTabs
}
