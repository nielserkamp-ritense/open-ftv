# Mock-data module

This Golang module contains the code to load and serve mock data in an extensible and flexible way.

## Sub-modules
- [Database schema](README_SCHEMA.md)
- [Transformations](README_TRANSFORMATIONS.md)
- [Table data](README_TABLE_DATA.md)
- [Custom endpoints](README_ENDPOINTS.md)
- [Filtering](README_FILTERING.md)
- [Field matching](README_FIELD_MATCHING.md)
- [Comparing](README_COMPARING.md)

## Run example.

```shell
git clone https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv.git
cd open-ftv
go run mock/datasources/generic/cmd/main.go --data=testdata/dataspaces/fds
```
In another shell:
```shell
curl -X POST -H'Content-Type: application/json' -H 'Accept: application/json' http://localhost:8443/v1/haalcentraal/api/brp/personen
```

For more, see the [generic service docs](../generic/README.md).


## License

[Licensed under the EUPL](../../LICENSE.md)
