# Mock-data module - transformations

[back to index](README.md)

## Transformations

Transformations can be used to "create" new output fields, using existing fields (or other transformations) as input.

This allows an endpoint to deliver "calculated" fields, which otherwise the calling application would have to itself.
The advantages are:
- It allows for very fine-grained data-minimalization.
- The calculation is in a central place, so calling applications do not need to re-invent the wheel.

Note that transformations happen after selection (vertical data-minimalization) but before field matching (horizontal data-minimalization).

Examples:
```yaml
      - id: "leeftijd"
        description: "Leeftijd van de persoon"
        transformationType: "age"
        resultType: "integer"
        isPII: true
        inputFields:
          "1": "geboortedatum"
      - id: "ouder-dan"
        description: "Indicatie of de persoon meer dan N jaar oud is"
        transformationType: "compare"
        compareType: "IsGreater"
        resultType: "bool"
        isPII: true
        inputTransformations:
          "1": "leeftijd"
        inputValues:
          "2": ":leeftijd:"
```

## Configuration:

- ```id``` - required - unique identifier of the transformation.
- ```description``` - optional - description of the transformation.
- ```transformationType``` - required - the type of transformation.
- ```compareType``` - required for ```compare``` transformation - the type of comparison to be performed.
- ```resultType``` - required for ```convert``` transformation - the data-type for the transformation result.
- ```isPII``` - optional - indicates the result data may contain *Persoonlijk Identificeerbare Informatie*.
- ```inputFields``` - optional - map of key/value pairs; the key defines the order of use; the value a database field identifier.
- ```inputTransforms``` - optional - map of key/value pairs; the key defines the order of use; the value a transformation identifier.
- ```inputValues``` - optional - map of key/value pairs; the key defines the order of use; the value a fixed value or a request parameter identifier [^1].

[^1] request parameter identifiers are indicated with a colon character at the start and end of their identifier; e.g., ```:param1:```.

## Transformation types

Currently, the following types of transformation are supported:
- ```convert```: performs data type conversions; e.g., from string to integer, or vice versa.
- ```compare```: compare a value with another value or list of values ([see comparing](README_COMPARING.md)).
- ```age```: determine the age in years given a date of birth.

The mechanism is very extensible, so it should be possible for other types of transformations to be added easily.

Future extensions could be:
- hashing a field,
- encrypting a field,
- masking a field,
- calculate the distance between two timestamps,
- etc... etc...

## Transformation result types

The result of a transformation can be converted to a specific data-type.

This is clear for the ```convert``` transformation type.

However, for ```compare``` and ```age``` it can also be used to:
- convert the result of the comparison from boolean to another type:
  - a string to contain ```true``` or ```false```.
  - an integer to contain ```1``` or ```0```.
- convert the result of the age calculation into a string.

## Transformation parameters

Transformations need input. This can come from various source data:
- data fields in the result set.
- other transformations.
- fixed value(s) in the transformation configuration.
- parameters sent in the request.

In the following example:
```yaml
      - id: "ouder-dan"
        description: "Indicatie of de persoon meer dan N jaar oud is"
        transformationType: "compare"
        compareType: "IsGreater"
        resultType: "bool"
        isPII: true
        inputTransformations:
          "1": "leeftijd"
        inputValues:
          "2": ":leeftijd:"
```
The first value for the comparison is another transformation with the unique identifier ```leeftijd```.
The second value is a request parameter (indicated with the colon characters) with the unique identifier ```leeftijd```.

In the following example:
```yaml
        inputTransformations:
          "1": "leeftijd"
        inputValues:
          "2": 18
```
Here the second parameter is the fixed numeric value ```18```.

## Request parameters

Data from the request can be used to supply transformations with necessary input value(s).

E.g., to determine if someone is older/younger than a particular age, the age to compare with can be supplied as a parameter.

Request parameters should be defined in a ```@params``` field in the URL query or the request body.

For a URL query string, the value of the ```@params``` field should be a fully escaped string with one or more key/value pairs separated by a comma character:
```
http://localhost:8080/v1/endpoint1?bsn=999999999,@params=leeftijd%3D18%2Cinkomen%3D30000
```
→ *leeftijd=18,inkomen=10000*

## Convert transformations

You can define transformations that convert a field with a specific data-type to another data-type.

A convert transformation must always have one input value.
It can be a data-field, another transformation, a fixed value, or even a request parameter.

The built-in conversions are based on common sense.
Impossible conversions (e.g., turning a structured field into a boolean) will fail.
Failed conversions will just return the default value for the resulting data-type.

Example:
```yaml
      - id: "inkomenX"
        description: "Inkomen als string"
        transformationType: "convert"
        resultType: "string"
        isPII: true
        inputFields:
          "1": "inkomen"
```
before:
```json
{"inkomen": 25100}
```
after:
```json
{"inkomen":25100,"inkomenX":"25100"}
```

## Compare transformations

A ```compare``` transformation must always have two input values.
These can be database fields, other transformations, fixed values, or even request parameters.

[See also: comparing](README_COMPARING.md).

## Age transformation

An ```age``` transformation must always have one input value.
This must be a valid date, timestamp, or a BRP style date (either or both month and day may be zero).

The result will be the age in years based on the difference between the current date and the given date.

---
[back to index](README.md)
