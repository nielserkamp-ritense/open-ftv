# Mock-data module - table data

[back to index](README.md)

## Table data formats

You can define table data as Comma Separated Values, YAML or JSON.

Please make sure not to use real names, addresses, phone numbers, email addresses or other personal information.

## CSV

By far the easiest way to add data to your tables.

Create a text file with a single header-line containing the identifiers of the fields you want to populate.
Add one or more lines (one for each row of data), with the field values in the same order as you added the field identifiers.

Example:
```csv
"bsn","straatnaam","huisnummer","huisletter","toevoeging","postcode","woonplaats"
"999990251","Brink","1","","rood","1111AA","Utrecht"
"999990263","Dorpsstraat","2","","II","1111BB","Amsterdam"
"999990275","Dorpslaan","3","C","","1111CC","Groningen"
"999990287","Uiteinde","4","","","1111DD","Rotterdam"
"999990299","Kerkstraat","5","","","1111EE","Utrecht"
"999990305","Kerklaan","6","","","1111FF","Groningen"
"999990317","Nieuwe Kerkstraat","7","B","1111GG","Rotterdam"
"999990329","Dorpsstraat","8","","3de verdieping","1111HH","Utrecht"
"999990330","Zomerlaan","9","","","1111II","Amsterdam"
"999990342","Winterlaan","101","A","","1111JJ","Amsterdam"
```

## YAML

Create an array of records, using field identifiers as keys:
```yaml
---
- bsn: "999990251"
  straatnaam: Brink
  huisnummer: 1
  huisletter: ""
  toevoeging: rood
  postcode: "1111AA"
  woonplaats: Utrecht
#   .
#   etc
#   .
```

## JSON

Create an array of records, using field identifiers as keys:
```json
[
  {"bsn":"999990251","straatnaam":"Brink","huisnummer":1,"toevoeging":"rood","postcode":"1111AA","woonplaats":"Utrecht"}
]
```

---
[back to index](README.md)
