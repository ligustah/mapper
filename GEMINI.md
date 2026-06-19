# Project Overview: Go Type Mapper

This project is a Go code generation tool designed to automate the creation of type mapping logic between different Go structs, including those generated from Protocol Buffers. It simplifies the process of converting data structures by allowing developers to define mapping interfaces with specific comments that guide the code generation.

**Key Technologies:**
*   **github.com/dave/jennifer:** Used for programmatic Go code generation.
*   **golang.org/x/tools/go/packages:** Utilized for parsing Go source code and extracting type information.

**Architecture:**
The tool operates in two main phases:
1.  **Parsing:** The `parser` component analyzes Go source files to identify interfaces annotated with `// mapper:generate`. It then inspects the methods within these interfaces to understand the desired type conversions and any field-specific mapping rules defined via comments (e.g., `// field:FieldName from:"SourceField"`).
2.  **Generation:** The `generator` component takes the parsed interface definitions and uses `dave/jennifer` to produce Go code that implements the specified mapping logic. This includes handling direct field assignments, pointer conversions, slice conversions, and custom field mappings.

## Building and Running

To build and run the code generator:

1.  **Build the project:**
    ```bash
    go build -o mapper .
    ```
2.  **Run the generator:**
    ```bash
    ./mapper
    ```
    This command will parse the `example` directory and generate the mapping code into the `example/gen` directory.

## Development Conventions

*   **Interface-driven Mapping:** Mapping logic is defined through Go interfaces.
*   **Doc Comment Annotations:**
    *   Interfaces to be processed by the generator must be annotated with `// mapper:generate`.
    *   Methods within these interfaces define the mapping functions.
    *   Field-level mapping rules can be specified using comments like `// field:OutputFieldName from:"InputFieldName"` or `// field:FieldName from:"-"` to ignore a field.
*   **Error Handling:** Mapping methods can return an error as a second return value, which the generator will incorporate into the generated code.
*   **Dependencies:** Project dependencies are managed via `go.mod`.
*   **Planning:** Present a clear and concise plan before proceeding with any multi-step implementation work.

## Tool Usage

For interacting with the Go codebase, the following tools are strongly preferred:

*   **gopls (Go Language Server):** For inquiries about the codebase, such as symbol definitions, references, and type information, `gopls` is the preferred tool. It provides accurate and comprehensive information by understanding the entire Go module.
    *   **Example `gopls` commands (via `go_search`, `go_symbol_references`, `go_file_context`, `go_package_api`, `go_diagnostics`):**
        *   `go_search(query="Mapper")`: Find all occurrences of the symbol "Mapper".
        *   `go_symbol_references(file="E:\code\mapper\example\mapper.go", symbol="Mapper")`: Find references to the `Mapper` interface.
        *   `go_file_context(file="E:\code\mapper\mapper\generator.go")`: Get a summary of cross-file dependencies for `generator.go`.
        *   `go_package_api(packagePaths=["github.com/ligustah/mapper/mapper"])`: Get API information for the `mapper` package.
        *   `go_diagnostics()`: Check for parse and build errors across the workspace.

*   **Delve Debugger:** For debugging, running tests, or executing the program, the `delve` debugger is the preferred tool. It allows for interactive debugging, setting breakpoints, inspecting variables, and stepping through code without the need for adding log lines.
    *   **Example `delve` commands (via `start_debug`, `set_breakpoint`, `continue`, `next`, `evaluate`, `list_local_vars`):**
        *   `start_debug(cwd="E:\code\mapper", program=".")`: Start debugging the main program.
        *   `start_debug(cwd="E:\code\mapper", program="E:\code\mapper\mapper", mode="test")`: Start debugging tests in the `mapper` package.
        *   `set_breakpoint(session_id="<session_id>", file="E:\code\mapper\mapper\generator.go", line=100)`: Set a breakpoint at a specific line.
        *   `continue(session_id="<session_id>")`: Continue execution until the next breakpoint or program end.
        *   `next(session_id="<session_id>")`: Step over the current line.
        *   `evaluate(session_id="<session_id>", expression="myVariable")`: Evaluate the value of `myVariable`.
        *   `list_local_vars(session_id="<session_id>")`: List local variables in the current scope.

These tools provide a powerful and efficient way to understand and debug the Go codebase, minimizing the need for manual file reading or intrusive logging.