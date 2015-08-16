AST usage
====

I'd really prefer to build AST's for each language, then output a stringified
representation of that - rather than just build strings that happen to be valid syntax.

I don't expect the string approach to be weak or short-term (languages do not often get rewritten, with old syntax invalidated), but it feels very wrong to be writing literals with newlines, tabs, brackets and variable names baked-in.

Repetition
====

Every generator looks very similar, and all approach the same goal of writing out equivalent code in slightly different syntax. But every one is essentially copy-pasted and modified from the last (with the JS one being the common ancestor).

The only way to avoid this is abstraction - making each generator a type, then defining methods which return the appropriate string for a given type. Unfortunately that makes it _hard_ to read, because the amount of methods would be huge. Every language has _slightly_ different structure, from semicolons to tabs to names of errors to quoting strategies to string concatenation to varible naming. Everything is different, it feels like it'd be _more_ confusing to abstract rather than just copy/paste changes across generators.
