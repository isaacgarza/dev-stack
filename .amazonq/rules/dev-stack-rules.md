# Instructions
* Please follow these instructions when generating code for me. If you do not follow these instructions, I will ask you to redo the changes.
* Give alternatives if applicable.
* Always ask me which option I want if there are multiple options.
* Always give me the pros and cons of each option if there are multiple options.
* Always give me a summary of the changes you made. But keep it brief.
* Always ask me if I want to add or expand on anything.
* Present the recommended option as the first option and give reasons why it is recommended at the end. I will choose the option I want or which I want to expand on. When I choose the option you can apply the changes.
* We want to prioritize clean, reusable, generic, extensible, and maintainable code that is production ready.
* We always want to be able to utilize the code for as many use cases as possible, never just the issue at hand. Never make changes that is special case and hardcoded for a single thing. Never suggest changes that have empty functions.
* Do not make assumptions about code that there is no context for. If there is any confusion or unknowns, reply with the unknowns and nothing else
* If the feature is large, break it down into smaller incremental changes. Give me the option to either proceed with the changes or generate docs that layout the plan for the feature changes. Always ask me if I want to proceed with the next change.
* Do not add comments to end of lines, always add comments above the line.
* Do not add docs unless I specifically ask for them. Give suggestions if applicable. Be as succinct as possible when generating docs or giving me summaries.
* Always prioritize existing code style and patterns in the project. If there is no existing code style or patterns, use industry best practices.
* Always prioritize simplicity over complexity. Do not over-engineer solutions.
* Always prioritize performance and efficiency. Do not suggest changes that will degrade performance or efficiency.

# Kotlin Projects
* @BeforeAll must be in a companion object at the top of the class also annotated with @JvmStatic
* Use shouldNotBeNull() instead of !!
* Use kotest assertions and matchers
* Use mockk instead of mockito
* Always use named arguements

# Java Projects

# Best Practices For All Projects
* If suggesting third-party libraries, prioritize maintained ones and be sure they are not deprecated as well as have been updated recently
* Never use wildcard imports (e.g., import java.util.*) and always sort imports
* For test classes, keep private functions at the top of the class after the property declarations, as well as companion objects
* Avoid mocking in integration tests as well as unit tests when possible
* Always name functions annotated with @BeforeAll, @BeforeEach, etc, the camel case name of the annotation
* Write unit tests for configuration classes
* Never use the annotation @ExtendWith(MockitoExtension::class); Set mocks at the field-level declaration when possible
* Avoid using `any()` when possible and prioritize specifying the type if using `any()` is required
* Don’t add comments to the code just to explain prompt changes. Comments to code should be production-ready and function level comments should be prioritized
* Do not make migration docs without first confirming if you need to make them
* Never add comments at the end of a line, always above the line
* Never use magic values. Always give them context. Pritoritze the use of constants, enums, or variables with context and use them in strings if applicable.
* Don’t make sweeping changes, do things incrementally
* Do not add tests that simply assert constant values
* Write tests the same way as existing tests in the project.
* Add kdocs or javadocs depending on the project similar to the other classes and functions in the project

# Go Projects
* Always use `context.Context` as the first parameter in functions that support cancellation, timeouts
* Always handle errors explicitly and avoid using panic
* Always check errors returned from functions and handle them appropriately
* Always use `defer` to close resources like files, database connections, etc.
* Always use `go fmt` to format code before committing
* Always use `go vet` to catch potential issues before committing
* Always use `golangci-lint` for linting and static analysis
* Always write unit tests for all functions and methods using the `testing` package
* Always use table-driven tests for better organization and readability
* Always use `testify` for assertions and mocking in tests
* Always write clear and concise comments for functions, methods, and complex logic
* Always use Go modules for dependency management
* Always prioritize standard library packages over third-party packages when possible
* Always handle nil values and avoid dereferencing nil pointers
* Always use channels and goroutines for concurrent programming when appropriate
* Always use interfaces to define behavior and promote decoupling
* Always use struct embedding to promote code reuse and composition
