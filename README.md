# restapi

## Motivation / Features

Off the shelf read-only REST API tool.
Works with csv, tsv, psv, json and jsond (josn dump) file types.
Works on windows, mac and linux.

It does everything it does from inside memory.

It won't give you the search speed or query flexibility of lucene or bleve.
At the same time, it it won't crap out when you add a small collection of only
a few million documents.

## Download

Prebuilt binaries can be downloaded from [here](https://github.com/nahidakbar/go-restapi/tree/master/bin).

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

## Data Types

### Comma Separated Values (CSV)

As CSVs are no longer the simple format they once were, be sure to save with cells quoted option enabled.

### Tab Separated Values (TSV)

Simple form assumed. I.e. newline separated rows; tab separated fields.

### Pipe Separated Values (PSV)

Simple form assumed. I.e. newline separated rows; pipe separated fields.

### JSON (.json)

Standard JSON array expected. All data types of a field are expected to be of the same type.

### JSON Dump (.jsond)

Newline separated JSON rows expected. All data types of a field are expected to be of the same type.


## Source Code

**WARNING: I hate lame tabs.**

    go get github.com/nahidakbar/go-restapi
