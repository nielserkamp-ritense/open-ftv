# Mock-data module - field matching (vertical data-minimalization)

[back to index](README.md)

## Query parameter

Define a query parameter in your request as follows:
```http://localhost:8080/v1/meta/tables?@fields=id,description```

E.g. **@fields=id,description**
The key "@fields" triggers the query parser to use the content of the parameter as a field matching expression.

## POST request as GET

If an endpoint is defined as a GET request, but is callable as a POST request,
the body of the request may contain data in the form of key/value pairs.

E.g., if the *Content-Type* header contains *application/json*, the content could look like this:
```json
{"x": "y", "@fields": "id,desc*"}
```

As you would expect, in such a case, our body parser would be able to determine that there is a parameter with the key *@fields*.
If so, the value of that parameter will be used as a field matching expression, as if it was defined in the query parameters of the request.

In the unlikely event that both the query parameters and the body contain a field matching expression,
the request handler will merge both expressions to form a single field matching expression.

## Field matching expressions

A field matching expression consists of one or more field names, separated by commas.

A field name may contain the following wildcard characters:
- an asterisk character ("*") means any character matches at this position.
- a question mark ("?") means a single character matches at this position.

Examples:
- ```*plaats``` will match any field with a unique identifier ending with the string ```plaats```.
- ```huis*``` will match any field with a unique identifier starting with the string ```huis```.
- ```b?n``` will match any field that is three characters long, starts with a ```b``` and ends with an ```n```.

It also supports qualified fields (including the table identifier).
E.g., ```persoon.*``` will match all fields from the ```persoon``` table.

As a result of the above rules, a single asterisk would match every field.
This is in fact the same when no field matching expression is found.
E.g., the default is to return all fields.

---
[back to index](README.md)
