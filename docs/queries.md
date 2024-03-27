### Page through my vehicles

```graphql
{
  vehicles(
    filterBy: {privileged: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}
    first: 10
  ) {
    edges {
      node {
        tokenId
        definition {
          make
          model
          year
        }
        owner
        mintedAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

Here we're asking for 10 cars to which `0xd8…045` has access. The default sort is descending by token id, so this query will produce the 10 most recently minted cars. For each of these cars, we are asking for the token id, owner, time of mint; and make, model, and year.

The parameter `first` and the elements `edges` and `pageInfo` come from the [Relay cursor spec](https://relay.dev/graphql/connections.htm). A response to this query might look like

```json
{
  "vehicles": {
    "edges": [
      {
        "node": {
          "tokenId": 130,
          "definition": {
            "make": "Lamborghini",
            "model": "Aventador",
            "year": 2022
          }
          "owner": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
          "mintedAt": "2023-05-01T09:00:12Z"
        }
      },
      // …
    ]
    "pageInfo": {
      "hasNextPage": true,
      "endCursor": "MTI5"
    },
    "totalCount": 22
  }
}
```

There are 22 total cars in this list and we're only showing 10. To see the next page, we modify the earlier query to use the `endCursor` from the first page:

```graphql
{
  vehicles(
    filterBy: {privileged: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}
    first: 10
    after: "MTI5"
  ) {
    # …
  }
}
```

### What's paired with my devices?

```graphql
{
  aftermarketDevices(
    filterBy: {owner: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}
    first: 10
  ) {
    edges {
      node {
        tokenId
        vehicle {
          tokenId
          owner
        }
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

Here we follow the link from device to attached vehicle, if any, to see the token id any attached vehicle as well as the owner.
