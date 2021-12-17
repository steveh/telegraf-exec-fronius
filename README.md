# telegraf-exec-fronius

This is a simple tool to extract [Fronius](https://www.fronius.com/en) solar data logger output and output
[Influx line protocol](https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/);
it is designed to be used with a
[telegraf exec plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/exec).

This parses the output of the Fronius HTTP API.

## Interactive Run Example

The compiled tool can be run interactively.

```bash
./telegraf-exec-fronius -help

Usage of telegraf-exec-fronius:
  -archive
    	Collect archive data
  -days uint
    	Days of history to collect (default 7)
  -host string
    	Fronius host (default "localhost")
  -inverter string
    	Collect inverter data with device ID (default "1")
  -meter string
    	Collect meter data with device ID (default "0")
  -realtime
    	Collect realtime data
  -system
    	Collect system data (default true)
```

## Telegraf Run Example

This is a sample telegraf exec input that assumes the binary has been installed
to `/usr/local/bin/telegraf-exec-fronius`:

```toml
[[inputs.exec]]
  commands = ["/usr/local/bin/telegraf-exec-fronius -host 10.0.0.10 -realtime"]
  timeout = "10s"
  data_format = "influx"

[[inputs.exec]]
  commands = ["/usr/local/bin/telegraf-exec-fronius -host 10.0.0.10 -archive -days 3"]
  interval = "1h"
  timeout = "60s"
  data_format = "influx"
```
