MAIN_SOURCE_FILE_DIRECTORY=./cmd/gen
EXECUTABLE_NAME=project-generator
EXECUTABLE_OUTPUT_DIRECTORY=./bin
EXECUTABLE=$(EXECUTABLE_OUTPUT_DIRECTORY)/$(EXECUTABLE_NAME)
GENERATION_FILE_LOCATION=./assets/project.gen
CONTENT_FILES_LOCATION=./assets/content-files
PROJECT_OUTPUT_LOCATION=./project-output

all: build

## Builds the executable and sets its permissions to allow the owner to execute it. Running just make will run this.
build:
	go build -o $(EXECUTABLE) -v $(MAIN_SOURCE_FILE_DIRECTORY)
	chmod 300 $(EXECUTABLE)

## Runs the previously build executable by first deleting the previous contents of the project output directory. Note that the output of rm here is ignored in case the project output directory was empty.
run:
	-rm -r $(PROJECT_OUTPUT_LOCATION)/*
	$(EXECUTABLE) -generation-file-location=$(GENERATION_FILE_LOCATION) -content-files-location=$(CONTENT_FILES_LOCATION) -project-output-location=$(PROJECT_OUTPUT_LOCATION)

## Deletes the executable and all project outputs created. Note that these rms as well have their outputs ignored for the case that they do not exist.
clean:
	-rm -r $(PROJECT_OUTPUT_LOCATION)/*
	-rm $(EXECUTABLE)
