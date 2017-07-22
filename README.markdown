# Sensors Bot

# Development

## Build

Cross-compile for Raspberry 3:

```bash
$ GOARM=7 GOARCH=arm GOOS=linux go build sensors.go
```

## Run

```bash
INFLUXDB_PASSWORD=bdf882e54c0fcb56ba25 ./bin/sensors --influxdb-database sandbox --influxdb-user sensors
```

## Iterate

Watch the source file; if changed, transfer it to the host `pi`:

```bash
fswatch *.go | xargs -I{} ./iterate.sh {}
```

# Deployment

```bash
ansible-playbook -i pi, deployment/playbook.yml
```

Ansible will also create the InfluxDB database, a R/O user for Grafana and a R/W user for the sensors bot.

## Troubleshooting

* Since the sensors bot is managed by systemd, the [usual ways](https://wiki.archlinux.org/index.php/Systemd#Troubleshooting) to inspect can be used.

    Examples:

    ```bash
    systemctl status sensors # get an overview
    journalctl --unit sensors.service -f # tail the logs
    ```

* If you try to read a value from the sensor and get a panic from Go, double-check that the device is registered with the framework (e.g. `[]gobot.Device{bme280}`).

* If you get an error `Humidity disabled`, simply restart the service. The sensor exposes this value only on after the first attempt to read it (check the data sheet).

# TODO

* Exit on humidity read failure so that systemd restarts the daemon
* Print the application version number at startup:
  - Use the output of `git rev-parse --short HEAD` to get the short commit SHA
  - Add a version variable to the go code that [needs to be set at build time](https://stackoverflow.com/a/11355611/3212907)- Check with `git diff-index --quiet HEAD -- sensors.go` whether the source file has uncommitted changes
  - If there are uncommitted changes, append something like '_dev-$(date "+%FT%T%z")' to the last commit SHA
  - set the variable at build time
