presilo
====

[![Build Status](https://travis-ci.org/Knetic/presilo.svg?branch=master)](https://travis-ci.org/Knetic/presilo)
[![Godoc](https://godoc.org/github.com/Knetic/presilo?status.png)](https://godoc.org/github.com/Knetic/presilo)

Generates (and updates) code for data models from a [JSON schema](http://json-schema.org/). Works for Go, C#, Java, Python, Lua, SQL, and Ruby.

Presilo is Esperanto for "printing press", because presilo makes it trivial to "reprint" the same content in different languages.

Why do I want this?
====

Before writing code, it's a good idea to define your data models. What fields is each structure going to have, the valid values for each field, and which fields need to be present for an instance to be valid. Doing this makes it more clear how your code will need to work, and lets you code your solutions much quicker and with less refactoring.

The best way I've found to do this is with a _schema_, which is a way to describe a data structure and its field.

Isn't there something else which does this?
====

There are a few projects which do similar things, but I don't find their patterns to match what I've found is the best method for defining schemas. I wanted one schema to be enforced on webservers (client request validation), client SDKs, and data persistence (DB schemas, file formats). One single schema that describes the data format for everything else. I wanted to generate large amounts of code for multiple languages, and not need to maintain any of it. I also wanted to be able to update existing code when a schema changes, without needing to manually edit code. I haven't found a project that does that.

There are some projects which take a piece of code, or a sql table definition, and try to generate other pieces of code or schemas. That's great if you fit exactly into those use cases. But I wanted to define a contract about the data which is _implemented_ by code or a DB schema. And, specifically, I didn't want a language- or DB-specific solution, I wanted to use this everywhere.

###Where's the binary?

This project is solely the library which handles parsing and codegen. You can reference this library in your own code, but most users will probably just want to head over to the [executable project](http://github.com/Knetic/presiloExecutable) and find a release there.

###Branching

I use green masters, and heavily develop with private feature branches. Full releases are pinned and unchangeable, representing the best available version with the best documentation and test coverage. Master branch, however, should always have all tests pass and implementations considered "working", even if it's just a first pass. Master should never panic.

###Activity

If this repository hasn't been updated in a while, it's probably because I don't have any outstanding issues to work on - it's not because I've abandoned the project. If you have questions, issues, or patches; I'm completely open to pull requests, issues opened on github, or emails from out of the blue.
