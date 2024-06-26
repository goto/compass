# Starring

Compass allows a user to stars an asset. This bookmarking functionality is introduced to increase the speed of a user to get information. 

To star and asset, we can use the User Starring API. Assuming we already have `asset_id` that we want to star.

```bash
$ curl --request PUT 'http://localhost:8080/v1beta1/me/starred/00c06ef7-badb-4236-9d9e-889697cbda46' \
--header 'Compass-User-UUID:gotocompany@email.com'
```

To get the list of my starred assets.
```bash
$ curl --request PUT 'http://localhost:8080/v1beta1/me/starred' \
--header 'Compass-User-UUID:gotocompany@email.com'

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
      }
  ]
}
```

There is also an API to see which users star an asset (stargazers) in the Asset API.

```bash
$ curl 'http://localhost:8080/v1beta1/assets/00c06ef7-badb-4236-9d9e-889697cbda46/stargazers' \
--header 'Compass-User-UUID:gotocompany@email.com'

{
  "data": [
      {
          "id": "1111-2222-3333",
          "email": "gotocompany@email.com",
          "provider": "shield"
      }
  ]
}
```