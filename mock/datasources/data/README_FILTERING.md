# Mock-data module - filtering (horizontal data-minimalization)

[back to index](README.md)

## Filter parameter

A filter parameter must be encoded in the URL query or the request body with the key ```@filter```.

The value should be a fully escaped string in JSON or YAML format or a simple text format.
The JSON and YAML format allow for complex filtering with the ```allOf``` and ```anyOf``` filter.

The top level element in YAML and JSON must be either an ```allOf``` or ```anyOf``` filter or a field value filter object (see below).

The text format supports multiple conditions separated by a comma character.
All conditions together are considered an ```allOf``` filter; e.g., all conditions must be met for a record to pass the filter.

Example YAML:
```yaml
- level: primary
  table: persoon
  field: bsn
  compare: exists
```

Example JSON:
```json
{"level": "primary", "table": "persoon", "field": "bsn", "compare": "exists"}
```

Example text:
```text
"bsn is not nil, postcode ~ ^(1111.*|2222.*)$"
```

## Field value filter

- ```level``` - required - the level at which the filter should be performed; allowed values are ```primary```, ```join```, ```any```.
- ```insensitive``` - optional - flag to indicate that string comparisons are to be performed case-insensitive (the default is case-sensitive).
- ```join``` - optional - unique identifier of a join the field to filter on is part of.
- ```table``` - optional - unique identifier of a table the field to filter on is a part of.
- ```field``` - required - the unique identifier of the field to filter on.
- ```compare``` - required - the type of comparison to perform.
- ```value``` - optional - value to compare the database field against.
- ```values``` - optional - list of values to compare the database field against (e.g., ```in list```, ```not in list```).

The ```value``` and ```values``` fields can be any data-type.
However, the data-type must be comparable to the data-type of the database field that is being compared.

Internally, the ```value``` and ```values``` field will be converted to the data-type of the database field before being compared.

## allOf filter

The ```allOf``` constructs requires all conditions within it to be true; comparable to an "AND" function.

Example:
```yaml
- allOf:
    - level: primary
      table: persoon
      field: bsn
      compare: exists
    - level: join
      join: adres
      field: postcode
      compare: "regex"
      value: "^(1111..|1112..)$"
```

```allOf``` and ```anyOf``` constructs can be nested.

## anyOf filter

The ```anyOf``` constructs requires at least one condition within it to be true; comparable to an "OR" function.

Example:
```yaml
- allOf:
    - level: join
      join: adres
      field: postcode
      compare: "wc"
      value: "1111*"
    - level: join
      join: adres
      field: postcode
      compare: "wc"
      value: "1112*"
```

```anyOf``` and ```allOf``` constructs can be nested.

---
[back to index](README.md)
