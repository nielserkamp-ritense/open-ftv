# Mock-data module - endpoints

[back to index](README.md)

## Built-in versus custom endpoints

By default, the system can serve the database schema as metadata from built-in endpoints.
Also, by default, the data in each table is accessible from a built-in endpoint.

Custom endpoints can be added through the configuration files.
These endpoints can join data from two or more tables, even recursively.
E.g., the BRP *persoon* table joins with the *ouder1* table, which in turn joins with (another row in) the *persoon* table.

For the built-in endpoints, an [OAS v3 spec](../../../oas/schema/openapi.yaml) is available.

## Custom endpoint definition

Example:
```yaml
---
- id: brp-personen
  version: 1
  type: GET
  calledAs: POST
  path: /haalcentraal/api/brp/personen
  fullVersion: 1.0.0
  description: HaalCentraal API
  datasource: brp
  table: persoon
  joins:
    - type: OptionalParentChild
      source: ouder1
      joins:
        - type: OptionalSibling
          source: persoon
          fields: ["bsn-ouder:bsn"]
    - type: OptionalParentChild
      source: ouder2
      joins:
        - type: OptionalSibling
          source: persoon
          fields: ["bsn-ouder:bsn"]
    - type: OptionalParentChild
      source: nationaliteit
#     .
#     etc.
#     .
```

### Endpoint definition ###

- id - *required* - unique identification of the endpoint.
- version - *required* - major version of the endpoint.
- type - *required* - http method indication the operation.
- calledAs - *optional* - overriding http method indicating how the endpoint will be called.
- path - *required* - the path of the endpoint.
- fullVersion - *required* - returned as the value for the "api-version" header in responses.
- description - *optional* - description of the endpoint.
- datasource - *required* - the datasource the endpoint will operate for.
- table - *required* - the primary table to retrieve data from.
- joins - *optional* - list of (SQL-like) joins, to enrich the rows from the primary table with additional information.

### Join definition ###

- type - required - type of join to perform:
  - ForcedParentChild = join with a parent-child relationship; child data is added as a single field containing an array of objects.
  - OptionalParentChild = same as ForcedParentChild; if no child data is found, the field is not added to the parent row.
  - ForcedSibling = join as a sibling relationship; for a single child, the data is added as separate fields in the parent row;
    multiple child rows are ignored or may throw an error in the future.
  - OptionalSibling = same as ForcedSibling; however, if no child data is found, no fields are added to the parent row.
- source - required - the id of the table to join.
- joins - optional - an array of subjoins.

---
[back to index](README.md)
