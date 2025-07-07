# Mock-data module - comparing

[back to index](README.md)

## Data filtering

Data filtering can make use of the comparison functionality described here.

Note that this requires the use of the ```@filter``` parameter in either the query or the body of the request.
This is because field-filters from a URL query can only use the equality test.

[See also: filtering](README_FILTERING.md).

## Transformations

Transformations can use the comparison functionality described here.

The transformation type ```compare``` indicates this use, and,
requires the user to also configure one of the comparison types below in the transformation entity.

[See also: transformations](README_TRANSFORMATIONS.md).

## Comparison types

Comparison types are character sequences or (set of) keywords that define the type of comparison to be performed.

Comparison types are case-insensitive.

Full list of supported comparison types:
- test on existence [^1]: ```exists```, ```!nil```, ```not nil```, ```is not nil```, ```not empty```, ```is not empty```.
- test on non-existence [^1]: ```not exists```, ```!exists```, ```nil```, ```is nil```, ```empty```, ```is empty```.
- test on equality: ```==```, ```=```, ```eq```, ```equal```, ```is equal```.
- test on inequality: ```!=```, ```<>```, ```ne```, ```!eq```, ```not equal```, ```is not equal```.
- test on lesser: ```<```, ```lt```, ```!ge```, ```lesser```, ```lesser than```, ```is lesser```, ```is lesser than```,
  ```smaller```, ```smaller than```, ```is smaller```, ```is smaller than```.
- test on lesser or equal: ```<=```, ```=<```, ```le```, ```!gt```, ```lesser or equal```, ```lesser than or equal```,
  ```is lesser or equal```, ```is lesser than or equal```, ```smaller or equal```, ```smaller than or equal```,
  ```is smaller or equal```, ```is smaller than or equal```.
- test on greater: ```>```, ```gt```, ```!le```, ```greater```, ```greater than```, ```is greater```, ```is greater than```.
- test on greater or equal: ```>=```, ```=>```, ```ge```, ```!lt```, ```greater or equal```, ```greater than or equal```,
  ```is greater or equal```, ```is greater than or equal```.
- test on existence in a list: ```in```, ```in list```.
- test on non-existence in a list: !```in```, ```!in list```, ```not in```, ```not in list```.
- test on equality with wildcard [^2][^3]: ```wc```, ```wildcard```.
- test on inequality with wildcard [^2][^3]: ```!wc```, ```!wildcard```, ```not wildcard```.
- test on equality with SQL style like [^3]: ```like```, ```is like```.
- test on inequality with SQL style like [^3]: ```!like```, ```not like```, ```is not like```.
- match on regular expression [^4]: ```~```, ```rx```, ```regex```, ```match rx```, ```match regex```.
- not match on regular expression [^4]: ```!~```, ```!rx```, ```!regex```, ```not rx```, ```not regex```, ```not match rx```, ```not match regex```.

[^1] this is exactly the same as ```IS NULL``` and ```IS NOT NULL``` in ANSI SQL.\
[^2] wildcard characters are:\
     - ```*```: zero, one or more of any character.\
     - ```?```: any character, exactly once.\
[^3] wildcard and like comparisons are internally converted to a regular expression match.\
[^4] the regular expression syntax is defined by the [Golang implementation](https://github.com/google/re2/wiki/Syntax).

---
[back to index](README.md)
