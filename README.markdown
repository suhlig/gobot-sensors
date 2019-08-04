# Sensors Bot

# Development

## Build

Cross-compile for Raspberry 3:

```bash
$ GOARM=7 GOARCH=arm GOOS=linux go build sensors.go
```

## Run

```bash
INFLUXDB_PASSWORD=S3CRET ./bin/sensors --influxdb-database localhost --influxdb-user sensors
```

## Iterate

Watch the source file; if changed, transfer it to the host `pi`:

```bash
fswatch *.go | xargs -I{} ./iterate.sh {}
```

# Deployment

```bash
$ ansible-playbook deployment/playbook.yml
```

Ansible will deploy the service and create a user with write privileges for the sensors bot.

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

* Cross-compile before deployment
* Determine the SHA
  - Use the output of `git rev-parse --short HEAD` to get the short commit SHA
  - If `git status --porcelain` is empty, the workspace is clean
  - If not clean, append something like '_dev-$(date "+%FT%T%z")' to the last commit SHA
* Inject the version into the go binary [at build time](https://stackoverflow.com/a/11355611/3212907)
* Print the application version number at startup
* Exit on humidity read failure so that systemd restarts the daemon
