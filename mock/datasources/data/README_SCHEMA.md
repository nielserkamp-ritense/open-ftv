# Mock-data module - database schema and data

[back to index](README.md)

## Loading the schema

The database schema is be loaded during startup from YAML or JSON encoded files.
If an error is encountered during loading or parsing of the schema files,
an error is returned and processing stops.

An initial *metadata.yaml* file must describe which elements will be loaded from where.

You can define one or more datasources, each with one or more tables and one or more custom endpoints.
Optionally, you can define a dataspace element, which encompasses all the datasources.

You can mix and match as you like:
- A single table in a single datasource.
- A few tables in a single datasource.
- Multiple datasources, each with one or a few tables.
- A top-level dataspace, with multiple datasources, each with one or more tables.

### metadata.yaml

The metadata file is used to get the location of the schema elements to be loaded into memory.

Example:
```yaml
---
dataspace: fds.yaml
datasource:
  - brp/schema.yaml
endpoints:
  - brp/endpoints.yaml
sourceData:
  brp:
    adres: brp/adres.csv
    afnemer: brp/afnemer.csv
    contact: brp/contact.csv
#     .
#     etc.
#     .
```

In this example, an FDS dataspace is defined along with a BRP datasource.

Also, custom endpoints will be loaded from the *brp/endpoints.yaml* file.

Table data is loaded from the given data files,
where the key of a file must match the ID of a defined table the data belongs to.

### Dataspace

A dataspace definition is for decoration. Only its ID is used in fully qualified names.

An example:
```yaml
---
id: FDS
description: Federatief Data Stelsel
```

### Datasources

Datasource defines a "bron" (Dutch) with any table definitions the datasource is built with.

(Shortened) Example:
```yaml
---
id: "brp"
description: "Basis Registratie Personen"
tables:
  - id: "persoon"
    description: "01/51 Persoonsgegevens"
    fields:
#     .
#     etc.
#     .
```

The datasource has a unique ID and optional description, and may contain one or more table definitions.

### Tables

```yaml
  - id: "persoon"
    description: "01/51 Persoonsgegevens"
    fields:
      #     .
      #     etc.
      #     .
    transformations:
      - id: "leeftijd"
        description: "Leeftijd van de persoon"
        transformationType: "age"
        resultType: "integer"
        isPII: true
        inputFields:
          "1": "geboortedatum"
    primaryKey:
      id: "pk"
      description: "Primary key: bsn"
      fields: ["bsn"]
    secondaryIndexes:
      - id: "ix2"
        description: "Index 2: a-nummer"
        fields: ["a-nummer"]
#     .
#     etc.
#     .
```

Each table definition has an ID and optional description, and may contain:
- one or more fields,
- one or more transformations,
- a primary key,
- one or more secondary indexes,
- one or more foreign keys.

### Fields

```yaml
      - id: "bsn"
        description: "01.20 Burgerservicenummer"
        type: "string"
        isPII: true
        minLen: 9
        maxLen: 9
        minValue: "000000000"
        maxValue: "999999999"
      - id: "voornamen"
        description: "02.10 Voornamen"
        type: "string"
        isPII: true
        minLen: 1
        maxLen: 200
```

Each table field has a unique ID, an optional description and a type,
and may contain one or more of the following constraints:
- a flag indicating if it is an array,
- a flag indicating if it is an enumeration,
- a flag indicating if the field contains "Persoonlijk Identificeerbare Informatie",
- a formatting expression,
- a minimum length,
- a maximum length,
- a minimum value,
- a maximum value,
- a list of allowed values.

### Transformations

[See here](README_TRANSFORMATIONS.md).

### Primary key and secondary indexes

### Foreign keys

### Endpoints

[See here](README_ENDPOINTS.md).

## Access to the schema

By default, the system can serve the database schema as metadata from built-in endpoints.
See the [OAS v3 spec](../../../oas/schema/openapi.yaml).

---
[back to index](README.md)
