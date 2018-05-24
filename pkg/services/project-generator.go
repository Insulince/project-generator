package services

import (
	"experimental/project-generator/pkg/configuration"
	"github.com/golang-collections/collections/stack"
	"fmt"
	"strings"
	"errors"
	"os"
	"log"
	"io/ioutil"
)

var config *configuration.Config

func GenerateProject(passedConfig *configuration.Config) (err error) {
	config = passedConfig

	// Parse the variables file into a map[string]string of variableName -> variableValue.
	variables, err := ParseVariablesFile()
	if err != nil {
		return err
	}

	// Parse the structure file into both generic structure data (variables left in) and specific structure data (variables swapped out for corresponding values).
	genericStructureData, specificStructureData, err := ParseStructureFile(variables)
	if err != nil {
		return err
	}

	// Parse the content file into a map[string]string of fileName -> fileContents.
	contentMap, err := ParseContentFile(variables)
	if err != nil {
		return err
	}

	fmt.Println(variables)
	fmt.Println("============================================================================================================================================")
	fmt.Println(genericStructureData)
	fmt.Println("============================================================================================================================================")
	fmt.Println(specificStructureData)
	fmt.Println("============================================================================================================================================")
	fmt.Println(contentMap)

	// Generate the project based on the structure and content map.
	err = BuildProject(genericStructureData, specificStructureData, contentMap)
	if err != nil {
		return err
	}

	return nil
}

func ParseVariablesFile() (variables map[string]string, err error) {
	// Get variables file data.
	variablesData, err := GetFileContents(config.VariablesFileLocation)
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
	genericStructureData, err = GetFileContents(config.StructureFileLocation)
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
		interpolatedVariableName := configuration.VariableInterpolationSymbol + variableName + configuration.VariableInterpolationSymbol
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

func ParseContentFile(variables map[string]string) (contentMap map[string]string, err error) {
	// Get the content data from the content file.
	contentData, err := GetFileContents(config.ContentFileLocation)
	if err != nil {
		return nil, err
	}

	// Build a map[string]string of generic file name -> specific file content.
	contentMap, err = BuildContentMap(contentData)
	if err != nil {
		return nil, err
	}

	// Substitute interpolated variable names with variable values in each entry of the contentMap.
	contentMap = SubstituteVariablesInContentMap(contentMap, variables)

	return contentMap, nil
}

func BuildContentMap(contentData string) (contentMap map[string]string, err error) {
	contentMap = make(map[string]string)

	contentSegments := strings.Split(contentData, configuration.ContentSegmentSeparator+"\n")
	for _, contentSegment := range contentSegments {
		if len(contentSegment) == 0 {
			continue // Then this is a blank line, skip (probably last line in file).
		}
		contentSegmentIdentifier := contentSegment[:strings.Index(contentSegment, "\n")]
		contentSegmentData := contentSegment[strings.Index(contentSegment, "\n")+1:]

		contentMap[contentSegmentIdentifier] = contentSegmentData
	}

	return contentMap, nil
}

func SubstituteVariablesInContentMap(contentMap map[string]string, variables map[string]string) (replacedContentMap map[string]string) {
	for variableName, variableValue := range variables {
		interpolatedVariableName := configuration.VariableInterpolationSymbol + variableName + configuration.VariableInterpolationSymbol
		for fileName, content := range contentMap {
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
			contentMap[fileName] = replacedContent
		}
	}
	replacedContentMap = contentMap

	return replacedContentMap
}

func BuildProject(genericStructureData string, specificStructureData string, contentMap map[string]string) (err error) {
	genericStructureDataLines := strings.Split(genericStructureData, "\n")
	specificStructureDataLines := strings.Split(specificStructureData, "\n")

	currentGenericPathStack := stack.New()

	currentSpecificPathStack := stack.New()
	currentSpecificPathStack.Push(config.OutputDirectoryLocation + "/")

	previousTabs := 0
	previousDirectoryTabs := 0
	for specificStructureDataLineIndex, specificStructureDataLine := range specificStructureDataLines {
		genericStructureDataLine := genericStructureDataLines[specificStructureDataLineIndex]

		if len(specificStructureDataLine) == 0 {
			continue // This is a blank line, skip (probably last line in file).
		} else {
			currentTabs := GetTabs(specificStructureDataLine)
			if strings.Index(specificStructureDataLine, configuration.DirectoryIndicatorSymbol) != -1 {
				genericDirectoryName := genericStructureDataLine[strings.LastIndex(genericStructureDataLine, "\t")+1:]
				specificDirectoryName := specificStructureDataLine[strings.LastIndex(specificStructureDataLine, "\t")+1:strings.Index(specificStructureDataLine, configuration.DirectoryIndicatorSymbol)] + "/"

				if currentTabs <= previousTabs {
					for currentTabs < previousTabs {
						currentGenericPathStack.Pop()
						currentSpecificPathStack.Pop()

						previousTabs--
					}

					currentGenericPathStack.Push(genericDirectoryName)
					currentSpecificPathStack.Push(specificDirectoryName)
				} else if currentTabs == previousTabs+1 {
					currentGenericPathStack.Push(genericDirectoryName)
					currentSpecificPathStack.Push(specificDirectoryName)
				} else {
					return errors.New("You cannot create a directory that is deep than the previous directory by more than 1 level. Every directory must have a direct parent/be a direct descendant. Check your gen file.")
				}

				specificDirectoryLocation := GetRelativePath(*currentSpecificPathStack)
				MakeDirectory(specificDirectoryLocation)

				previousDirectoryTabs = currentTabs
			} else {
				genericFileName := genericStructureDataLine[strings.LastIndex(genericStructureDataLine, "\t")+1:]
				specificFileName := specificStructureDataLine[strings.LastIndex(specificStructureDataLine, "\t")+1:]

				if currentTabs > previousDirectoryTabs+1 {
					return errors.New("You cannot create a file that is deeper than the previous directory by more than 1 level. Every file must have a direct parent/be a direct descendant. Check your gen file.")
				}

				if previousTabs > currentTabs {
					for previousTabs > currentTabs {
						currentGenericPathStack.Pop()
						currentSpecificPathStack.Pop()

						previousTabs--
						previousDirectoryTabs--
					}
				}

				specificFileLocation := GetRelativePath(*currentSpecificPathStack) + specificFileName
				genericFileLocation := GetRelativePath(*currentGenericPathStack) + genericFileName
				MakeFile(specificFileLocation, contentMap[genericFileLocation])
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
	fmt.Printf("Making file\t\t\"%v\"\n", location)
	if err := ioutil.WriteFile(location, []byte(content), 0777); err != nil {
		log.Fatalf(err.Error())
	}
}

func GetTabs(str string) (numTabs int) {
	for _, v := range str {
		if string(v) == "\t" {
			numTabs++
		}
	}

	return numTabs
}

func GetFileContents(fileLocation string) (fileContents string, err error) {
	rawFileContents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return "", err
	}

	fileContents = string(rawFileContents)

	return fileContents, nil
}
