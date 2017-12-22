# go-whosonfirst-lieu

## Important

Stop. This is too soon for you.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Tools

### lieu-prepare-atp

_Please write me_

### lieu-prepare-wof

```
./bin/lieu-prepare-wof -timings -processes 100 -out /usr/local/data/whosonfirst-work/lieu.geojson -mode sqlite /usr/local/data/whosonfirst-sqlite/*venue*.db
2017/12/22 17:07:39 time to prepare /usr/local/data/whosonfirst-sqlite/whosonfirst-data-venue-ae-latest.db 4.538474ms
2017/12/22 17:07:39 time to prepare /usr/local/data/whosonfirst-sqlite/whosonfirst-data-venue-af-latest.db 1.932549ms
2017/12/22 17:07:39 time to prepare /usr/local/data/whosonfirst-sqlite/whosonfirst-data-venue-ar-latest.db 6.987304ms
2017/12/22 17:08:09 time to process 43501 records 30.000394412s
2017/12/22 17:08:39 time to process 88095 records 1m0.000648475s
...time passes...
```

## See also