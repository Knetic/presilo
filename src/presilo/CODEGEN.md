Code-generation contract
====

Every language follows the same contract when it comes to generating type declarations and methods.
That contract is listed here.

Type Declarations
====

Only object schemas will be given type declarations. Numbers, integers, strings, bools, and arrays
are considered children of object schemas.

A code generator must only handle one object schema at a time. It may be invoked multiple times, one for each object schema defined in the file; but no global state should be modified each run.

Constructors
====

Constructors are generated for each object type. The required arguments to any constructor are the same
as the object's "required" list of properties, and nothing else. If a required parameter has constraints,
then the constructor _may_ throw a language-appropriate error if the constraints are not met.

Languages without types will have built-in type checking, and throw the appropriate language error if
any argument doesn't match the expected type.

Getter/Setters
====

If a property has constraints, getter/setter methods will be generated for that property, and it will
be considered "protected" by the class. The setters will implement the field constraints.

If there are no constraints, no getter/setters will be generated, and consumers will be able to modify the field directly.

However, there is a user-defined setting called "-g" which means "generate getter/setter for all fields",
which if present, should be honored.

Serialization
====

In languages with built-in serialization/deserialization capabilities, both XML and JSON directives will be
added to the generated type. If the language has a standard json/xml hash library, the appropriate hash-to-object method will be created instead. If the language only has deserialization _libraries_, then no assumptions will be made, and no code generated.
