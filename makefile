MAIN_SOURCE_FILE_DIRECTORY=./cmd/gen
EXECUTABLE_NAME=project-generator
EXECUTABLE_OUTPUT_DIRECTORY=./bin
EXECUTABLE=$(EXECUTABLE_OUTPUT_DIRECTORY)/$(EXECUTABLE_NAME)
VARIABLES_FILE_LOCATION=./specification/variables.pgen
STRUCTURE_FILE_LOCATION=./specification/structure.pgen
CONTENT_FILE_LOCATION=./specification/content.pgen
OUTPUT_DIRECTORY_LOCATION=./out

## Builds the executable and sets its permissions to allow the owner to execute it. Running just make will run this.
build:
	go build -o $(EXECUTABLE) -v $(MAIN_SOURCE_FILE_DIRECTORY)
	chmod 300 $(EXECUTABLE)

## Runs a previously built executable.
run:
	$(EXECUTABLE) -v=$(VARIABLES_FILE_LOCATION) -s=$(STRUCTURE_FILE_LOCATION) -c=$(CONTENT_FILE_LOCATION) -o=$(OUTPUT_DIRECTORY_LOCATION)

## Deletes the executable and all project outputs created. Note that these rms as well have their outputs ignored for the case that they do not exist.
clean:
	-rm -r $(OUTPUT_DIRECTORY_LOCATION)/*
	-rm $(EXECUTABLE)

## A custom target used for IntelliJ's "Makefile" plugin to connect to.
intellij: clean	build run