# fah-collector

Data collector that works with [fah-sidecar](https://github.com/ebiiim/fah-sidecar) by collecting FAH data.

## Usage

### Version 2

Requirements: `fah-sidecar` versions in [1.1, ]

#### Endpoint for Sidecars: POST HTTP(S)://{HOST}
Every fah-sidecar instance sends data to this URL. Uses `sc_hostname` in data as identifier.

NOTE: No need to pass identifier via URL anymore. fah-sidecar v1.1.0 contains `sc_nodename` and `sc_hostname` for identification.

### Endpoint for Viewers: GET HTTP(S)://{HOST}/all
Get current data in JSON.
- Hold 15 secs if the data with a particular identifier has not been updated, then drop it.
- Cache the response JSON for 1 secs by default. `./fah-collector ttl=1`


### Version 1

Requirements: `fah-sidecar` versions in [1.0, 1.1]

#### Endpoint for Sidecars: POST HTTP(S)://{HOST}/{IDENTIFIER}
Every fah-sidecar instance sends data to this URL with its identifier (e.g., hostname, Pod name).

### Endpoint for Viewers: GET HTTP(S)://{HOST}/all
Get current data in JSON.
- Hold 15 secs if the data with a particular identifier has not been updated, then drop it.
- Cache the response JSON for 5 secs by default. `./fah-collector ttl=5`
