import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Querying  metadata

## Prerequisites

This guide assumes that you have a local instance of compass running and listening on `localhost:8080`. See [Installation](installation.md) guide for information on how to run Compass.
## Using the Search API

#### We can search for required text in the following ways:

1. Using **`compass search <text>`** CLI command
2. Calling to **`GET /v1beta1/search`** API with `text` to be searched as query parameter

The API contract is available [here](https://github.com/goto/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

To demonstrate how to use compass, we’re going to query it for resources that contain the word ‘booking’.

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

This will return a list of search results. Here’s a sample response:

```bash
{
    "data": [
        {
            "id": "00c06ef7-badb-4236-9d9e-889697cbda46",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-seg-enriched-booking-dagger",
            "type": "topic",
            "service": "kafka",
            "name": "g-godata-id-seg-enriched-booking-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        },
        {
            "id": "9e69c08a-c3c2-4e04-957f-c8010c1e6515",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-booking-bach-test-dagger",
            "type": "topic",
            "service": "kafka",
            "name": "g-godata-id-booking-bach-test-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        },
        {
            "id": "ff597a0f-8062-4370-a54c-fd6f6c12d2a0",
            "urn": "kafka::g-godata-id-playground/ g-godata-id-booking-bach-test-3-dagger",
            "type": "topic",
            "service": "kafka",
            "title": "g-godata-id-booking-bach-test-3-dagger",
            "description": "",
            "labels": {
                "flink_name": "g-godata-id-playground",
                "sink_type": "kafka"
            }
        }
    ]
}
```

Compass decouple identifier from external system with the one that is being used internally. ID is the internally auto-generated unique identifier. URN is the external identifier of the asset, while Name is the human friendly name for it. See the complete API spec to learn more about what the rest of the fields mean.

### Filter

Compass search supports restricting search results via filter by passing it in query params. Filter query params format is **`filter[{field_key}]={value}`** where **`field_key`** is the field name that we want to restrict and **`value`** is what value that should be matched. Filter can also support nested field by chaining key **`field_key`** with **`.`** \(dot\) such as **`filter[{field_key}.{nested_field_key}]={value}`**. 
#### We can filter our search in the following ways:

1. Using **`compass search <text> --filter=field_key1:val1`** CLI command
2. Calling to **`GET /v1beta1/search`** API with **`text`** and **`filter[field_key1]=val1`** as query parameters 

For instance, to restrict search results to the ‘id’ landscape for ‘gotocompany’ organisation, run:

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking --filter=labels.landscape=id,labels.entity=gotocompany
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking&filter[labels.landscape]=id&filter[labels.entity]=gotocompany' \
--header 'Compass-User-UUID:gotocompany@email.com'
```
</TabItem>
</Tabs>

Under the hood, filter's work by checking whether the matching document's contain the filter key and checking if their values match. Filters can be specified multiple times to specify a set of filter criteria. For example, to search for ‘booking’ in both ‘vn’ and ‘th’ landscape, run:

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking --filter=labels.landscape=vn,labels.landscape=th
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking&filter[labels.landscape]=vn&filter[labels.landscape]=th' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

### Query

Apart from filters, Compass search API also supports fuzzy restriction in its query params. The difference of filter and query are, filter is for exact match on a specific field in asset while query is for fuzzy match.
#### We can search with custom queries in the following ways:

1. Using **`compass search <text> --query=field_key1:val1`** CLI command
2. Calling to **`GET /v1beta1/search`** API with **`text`** and **`query[field_key1]=val1`** as query parameters 

Query format is not different with filter `query[{field_key}]={value}` where `field_key` is the field name that we want to query and `value` is what value that should be fuzzy matched. Query could also support nested field by chaining key `field_key` with `.` \(dot\) such as `query[{field_key}.{nested_field_key}]={value}`. For instance, to search results that has a name `kafka` and belongs to the team `data_engineering`, run:

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking --query=name:kafka,labels.team=data_eng
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking&query[name]=kafka&query[labels.team]=data_eng' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

### Ranking Results
Compass allows user to rank the results based on a numeric field in the asset. It supports nested field by using the `.` \(dot\) to point to the nested field. For instance, to rank the search results based on `usage_count` in `data` field, run:

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking --rankby=data.usage_count
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking&rankby=data.usage_count' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

### Size
You can also specify the number of maximum results you want compass to return using the **`size`** parameter

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass search booking --size=5
```
</TabItem>
<TabItem value="HTTP" label="HTTP">

```bash
$ curl 'http://localhost:8080/v1beta1/search?text=booking&size=5' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

## Using the Suggest API
The Suggest API gives a number of suggestion based on asset's name. There are 5 suggestions by default return by this API.

The API contract is available [here](https://github.com/goto/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

Example of searching assets suggestion that has a name ‘booking’.

```bash
$ curl 'http://localhost:8080/v1beta1/search/suggest?text=booking' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
This will return a list of suggestions. Here’s a sample response:

```bash
{
    "data": [
        "booking-daily-test-962ZFY",
        "booking-daily-test-c7OUZv",
        "booking-weekly-test-fmDeUf",
        "booking-daily-test-jkQS2b",
        "booking-daily-test-m6Oe9M"
    ]
}
```
## Using the Get Assets API
The Get Assets API returns assets from Compass' main storage (PostgreSQL) while the Search API returns assets from Elasticsearch. The Get Assets API has several options (filters, size, offset, etc...) in its query params.


|  Query Params | Description |
|---|---|
|`types=topic,table`| filter by types |
|`services=kafka,postgres`| filter by services |
|`data[dataset]=booking&data[project]=p-godata-id`| filter by field in asset.data |
|`q=internal&q_fields=name,urn,description,services`| querying by field|
|`sort=created_at`|sort by certain fields|
|`direction=desc`|sorting direction (asc / desc)|


The API contract is available [here](https://github.com/goto/compass/blob/main/third_party/OpenAPI/compass.swagger.json).

## Using the Lineage API

The Lineage API allows the clients to query the data flow relationship between different assets managed by Compass.

See the swagger definition of [Lineage API](https://github.com/goto/compass/blob/main/third_party/OpenAPI/compass.swagger.json)) for more information.

Lineage API returns a list of directed edges. For each edge, there are `source` and `target` fields that represent nodes to indicate the direction of the edge. Each edge could have an optional property in the `prop` field.

#### We can search for lineage in the following ways:

1. Using **`compass lineage <urn>`** CLI command
2. Calling to **`GET /v1beta1/lineage/:urn`** API with `urn` to be searched as the path parameter

<Tabs groupId="cli" >
<TabItem value="CLI" label="CLI">

```bash
$ compass lineage data-project:datalake.events
```
</TabItem>
<TabItem value="HTTP" label="HTTP">


```bash
$ curl 'http://localhost:8080/v1beta1/lineage/data-project%3Adatalake.events' \
--header 'Compass-User-UUID:gotocompany@email.com' 
```
</TabItem>
</Tabs>

```json
{
    "data": [
        {
            "source": {
                "urn": "data-project:datalake.events",
                "type": "table",
                "service": "bigquery",
            },
            "target": {
                "urn": "events-transform-dwh",
                "type": "csv",
                "service": "s3",
            },
            "prop": {}
        },
        {
            "source": {
                "urn": "events-ingestion",
                "type": "topic",
                "service": "beast",
            },
            "target": {
                "urn": "data-project:datalake.events",
                "type": "table",
                "service": "bigquery",
            },
            "prop": {}
        },
    ]
}
```

The lineage is fetched from the perspective of an asset. The response shows it has a list of upstreams and downstreams assets of the requested asset.
Notice that in the URL, we are using `urn` instead of `id`. The reason is because we use `urn` as a main identifier in our lineage storage. We don't use `id` to store the lineage as a main identifier, because `id` is internally auto generated and in lineage, there might be some assets that we don't store in our Compass' storage yet.

