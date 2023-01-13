# Orderbook-Fetcher
A small tool I wrote to fetch orderbooks from the Eve ESI
and write them to csv files (as well as learn about golang html templates).

## API & Web Interface
The API for this tool listens on tcp port 8080 on all interfaces. \
and serves all files in the orderbooks directory. \
The files/orderbooks adhere to the following naming scheme:
```
/orderbooks/{<LOCATION>}_{TIMESTAMP}.csv
```
At ``/index.html`` there is a small web interface, showing the orderbooks that have been fetched
during the current session

## Refresh Token
To fetch market orders from citadels, as well their names, ESI authentication is required. \
Register an ESI application [here](https://developers.eveonline.com/) with the following scopes:
- esi-markets.structure_markets.v1
- esi-universe.read_structures.v1


## Configuration File
The configuration file contains the following options:
- retentionPeriod: Will keep the last n orderbooks per location. A value of zero will keep all orderbooks
- interval: Fetch every n orderbooks. Interval of 1 will fetch every orderbook
- regions: Fetches the orderbooks for those regions
- citadels: Fetches the orderbooks for those citadels
- clientId: (only required when fetching citadel orders) client id of the application that your character authed with
- refreshToken: (only required when fetching citadel orders) Refresh token for the authenticated character