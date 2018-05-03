# Project Generator
Build projects from custom or pre-made templates that are highly flexible and easily changeable.

### Concepts
**Project Generation Specification File** - This is a file which contains the structure of your desired project as a template that supports variable substitution. It is processed as follows:
1. For every line starting from the first line onwards, if this line is not an empty line, then this line should be considered a variable declaration.
2. Split this line based on the `=` character. Left of `=` is the name of the variable, right of `=` is the value.
3. Once we reach an empty line, advance one more line. Assume that this line is the root of your project and contains 0 tabs (tabs indicate structure).
4. Iterate over every line remaining in the project generation specification file searching for any mention of a variable by name, sandwiched between two `•`. So for example, if you had a variable `weight`. This step is searching every line for `•weight•`. Replace that with the actual value of the variable.
5. Go back to the first line after the blank line following the variable declarations.
6. For each line from here, check for the following:
    - Does this line contain `/` at the end?
        - Yes: This line represents a directory.
        - No: This line represents a file.
7. Count the number of tabs preceding the content of this line. Standard logic applies here to ensure that there is a tree-like structure being represented by the amount of tabs. Here is an example of such a structure with valid and invalid jumps in the number of tabs:
    ```
    foo/
        bar1/
            baz1.go
            baz2.go
        bar2/
                qux1/
                    baz3.go
        bar3/
            qux2/
                baz4.go
                    baz5.go
                baz6.go
        baz7.go
    ```
    This shows valid jumps in tabs. `foo` is the root of the project, `bar1`, `bar2`, `bar3`, and `baz7.go` all exist directly inside of `foo`. This also shows 2 invalid jumps. The jump from `bar2` to `qux1` can't be quantified because it suggests there should be some directory between the two. This would be rejected. Also the jump from `baz4.go` to `baz5.go` is invalid, for it suggests that `baz5.go` is a child of `baz4.go`, but both are files. This makes it somewhat tricky, because different rules apply for different types of lines.
8. Create the respective resource at the desired location.
9. Do not anticipate ever running into a line with 0 tabs or with no content. This should only happen in the root folder, and also this means you can't have any trailing new lines.

Also do note, you must use an actual tab, not just four spaces, this is the character searched for by the algorithm.

So a working example of a Go API named `wiget-api`:
```
name=widget-api
resource=widget

•name•/
	cmd/
		srv/
			main.go
			config.json
	pkg/
		configurations/
			config.go
		database/
			db.go
		handlers/
			health.go
			home.go
			not-found.go
			•resource•.go
		models/
			responses/
				error.go
				message.go
			api-request.go
			api-response-writer.go
			•resource•.go
		router/
			router.go
	.gitignore
	README.md
```

## Usage
### Install & Build
- Pull the repository into your machine: `git pull https://github.com/Insulince/project-generator.git`
- Enter the project: `cd ./project-generator`
- Build the executable: `make`
- The executable by default is created in the `./bin` folder.
### Running
- Run the executable manually:
    - Execute `./bin/project-generator`
    - Possible command line arguments are:
        - `generation-file-location` - Where your project generation specification file is located. Defaults to `./project.gen`
        - `content-files-location` - Where your content files will be located. Defaults to `./content-files`
        - `project-output-location` - Where the resulting project structure should be placed at. Defaults to `./project-output`
- Run the executable via `make`:
    - Configure the values of the command line flags in the `makefile`, or just accept the default values.
    - Execute `make run`
- Upon successful completion, your generated files will exist in whatever `project-output-location` was set to.

## Notes
- If you are using a lot of variables in your generation file, and it is the case that some variables are contained within others, you need to be careful of the ordering. You should place variables in order of largest variable to smallest to prevent a failure to replace it when looking for the content file corresponding to it. Example:
```
name=resource-xyz-v0.0.3
resource=xyz-v0.0.3
identifier=xyz
version=v0.0.3
```
This is the **correct** way to enter these variable. In the content-file building process, we scan starting from the first variable onwards for variables to replace in the file names based on their value. So if you wanted to replace the entire `reosurce-xyz-v0.0.3` with `•name•`, you have to put it first, otherwise portions of that file name would be replaced by the other variables matching to it first and once we finally do get to `•name•`, there would be no valid string left to match it. This feels like a horrible design choice and should probably be changed in the future, but I am just noting it for now.

## TODO
- Need to get a solid method for injecting the data in the content files into the generated files.
- Need to allow for variable substitutions within the content files.
- Templates!
- Allow for use of previously declared variables in newly declared variables. Ex:
    ```
    name=something          // something
    fullname=•name•-full    // something-full
    ```