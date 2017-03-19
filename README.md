# restapi

## CLI Usage

    ./restapi -address :8080 -datafile collection.json -path /api/v0/collection/

## REST API

### GET /schema.json

Returns general record schema.

### GET /id.json

Returns record by id. Ids run from 0 to N-1.

### GET /searchMeta.json

Returns search Metadata

### GET /search.json?queryList

Performs simple search and returns a list records.

Note that only simple query is supported. It performs simple O(n) search.
Put a cache in front if you need extra performance.

Query format is field=filter:value or field=value.

Query is &amp; joined. Space can usually also be replaced with +.

field=value means that the filter is the first listed in the metadata.

Search filter does not support relational and/or operators.

Quotes and - are supported.

See search metadata for a list of filters supported by fields.

Also note that only a limited number of results will be returned.

You can specify an offset parameter to get further results.
