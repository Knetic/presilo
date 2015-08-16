presilo
====

Generates (and updates) code for data models from a [JSON schema](http://json-schema.org/). Works for Go, C#, Java, Python, Lua, SQL, and Ruby.

Presilo is Esperanto for "printing press", because presilo makes it trivial to "reprint" the same content in different languages.

Why do I want this?
====

Before writing code, it's a good idea to define your data models. What fields is each structure going to have, the valid values for each field, and which fields need to be present for an instance to be valid. Doing this makes it more clear how your code will need to work, and lets you code your solutions much quicker and with less refactoring.

The best way I've found to do this is with a _schema_, which is a way to describe a data structure and its field.

How do I use it?
====

Start with a \*.json file whose contents describe your schema. See the [json-schema](http://json-schema.org/examples.html) examples for how to do this. Let's say our schema looks like this;

    {
    	"type": "object",
    	"title": "Person",
    	"properties":
    	{
    		"firstName":
    		{
    			"type": "string"
    		},
    		"lastName":
    		{
    			"type": "string"
    		},
    		"age":
    		{
    			"description": "Age in years",
    			"type": "integer",
    			"minimum": 0
    		}
    	},
    	"required": ["firstName", "lastName"]
    }

Then, run `presilo` on that file, and describe which language you'd like to have generated, and the output file.

    presilo -l go schema.json

This will generate a file called "Person.go". The name of the file is taken from the "title" field inside the schema itself. The contents of the generated file will contain a single struct definition, which ought to look like this;

    package main

    import (
        "errors"
    )

    type Person struct {

        firstName string
        lastName string
        age int
    }

    func NewPerson(firstName string, lastName string) *Person {

        var ret *schema

        ret = new(schema)
        ret.firstName = firstName
        ret.lastName = lastName

        return ret
    }

    func (this *Person) SetAge(value int) error {

        if(value < 0) {
            return errors.New("'age' must be greater than 0")
        }

        this.age = value
        return nil
    }

Sometimes your schema will have a nested schema, where one object contains another one. This will generate multiple files, one for each schema. Like so;

    {
        "title": "Car",
        "type": "object",
        "properties":
        {
            "seats"
            {
                "type": "integer",
                "minimum": "1",
                "maximum": "4"
            },
            "wheels":
            {
                "type": "array",
                "items":
                {
                    "oneOf":
                    [
                        {
                            "type": "object",
                            "title": "Tire",
                            "properties":
                            {
                                "weight":
                                {
                                    "type": "float",
                                    "minimum": "0"
                                },
                                "material":
                                {
                                    "type": "string",
                                    "enum": ["rubber", "unobtanium"]
                                }
                            }
                        }
                    ]
                }
            }
        }
    }

This will generate both Car.go and Tire.go, each containing the relevant data model and associated functions. The "Car" model will actually reference the "Tire" model. Note that if there are multiple schemas with the same title, an error will be given. You _must_ uniquely name your models.

That's cool, but not everyone uses Go. What about other languages? Using `-a` lists all supported languages.

    presilo -a

All supported languages have some form of "module" notion, which should be specified by the `-m` flag. For instance:

    presilo -l go -m awesome schema.json

Will make it so that instead of `package main`, the generated files will be listed as `package awesome`. Languages with nested package names (Java, Ruby, C#) will accept the dot-notation to separate package names (`com.awesome.project`, for example). If you give nested package names and tell `presilo` to generate a schema for a language which does not support them, it will give you an error.

Isn't there something else which does this?
====

There are a few projects which do similar things, but I don't find their patterns to match what I've found is the best method for defining schemas. I wanted one schema to be enforced on webservers (client request validation), client SDKs, and data persistence (DB schemas, file formats). One single schema that describes the data format for everything else. I wanted to generate large amounts of code for multiple languages, and not need to maintain any of it. I also wanted to be able to update existing code when a schema changes, without needing to manually edit code. I haven't found a project that does that.

There are some projects which take a piece of code, or a sql table definition, and try to generate other pieces of code or schemas. That's great if you fit exactly into those use cases. But I wanted to define a contract about the data which is _implemented_ by code or a DB schema. And, specifically, I didn't want a language- or DB-specific solution, I wanted to use this everywhere.
