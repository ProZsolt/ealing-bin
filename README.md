# Ealing Bin

A small webservice to quikly look up which bin is collected next time in Ealing

## Install

```
go install github.com/prozsolt/ealing-bin
```

## Usage

Look up the UPRN corresponding to your address

```
ealing-bin addresses W3 9JW
12103517: 1 HEREFORD ROAD, ACTON, LONDON, W3 9JW
12103518: 10 HEREFORD ROAD, ACTON, LONDON, W3 9JW
12103519: 11 HEREFORD ROAD, ACTON, LONDON, W3 9JW
12103520: 12 HEREFORD ROAD, ACTON, LONDON, W3 9JW
12103521: 13 HEREFORD ROAD, ACTON, LONDON, W3 9JW
...
```

Set the `EALING_BIN_UPRN` enviroment variable to the UPRN corresponding to your address

Run th web server `ealing-bin serve`

Open your browser `http://localhost:8080/`

## Limitation

Some APIs are not reliable and only tested on my own address.