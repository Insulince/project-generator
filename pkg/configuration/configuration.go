package configuration

import (
	"log"
	"flag"
)

const DefaultVariablesFileLocation = "./specification/variables.pgen"
const DefaultStructureFileLocation = "./specification/structure.pgen"
const DefaultContentFileLocation = "./specification/content.pgen"
const DefaultOutputDirectoryLocation = "./out"

const VariableInterpolationSymbol = "•"
const ContentSegmentSeparator = VariableInterpolationSymbol + VariableInterpolationSymbol + VariableInterpolationSymbol
const DirectoryIndicatorSymbol = "|"

type Config struct {
	VariablesFileLocation   string `json:"variablesFileLocation"`
	StructureFileLocation   string `json:"structureFileLocation"`
	ContentFileLocation     string `json:"contentFileLocation"`
	OutputDirectoryLocation string `json:"outputDirectoryLocation"`
}

//LoadConfig gets the configuration values for the api
func LoadConfig() (config *Config, err error) {
	config = &Config{}

	config.VariablesFileLocation = *flag.String("v", DefaultVariablesFileLocation, "The location of the project variables file. Default: "+DefaultVariablesFileLocation)
	config.StructureFileLocation = *flag.String("s", DefaultStructureFileLocation, "The location of the project structure file. Default: "+DefaultStructureFileLocation)
	config.ContentFileLocation = *flag.String("c", DefaultContentFileLocation, "The location to of the project content file. Default: "+DefaultContentFileLocation)
	config.OutputDirectoryLocation = *flag.String("o", DefaultOutputDirectoryLocation, "The location to output directory for the generated project. Default: "+DefaultOutputDirectoryLocation)
	flag.Parse()

	log.Printf("Config successfully loaded.\n")
	return config, nil
}
