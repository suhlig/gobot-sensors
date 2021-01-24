# Sensors Bot

# Development

## Build

Cross-compile for Raspberry 3:

```bash
$ GOARM=7 GOARCH=arm GOOS=linux go build sensors.go
```

## Run

```bash
./sensors -influxdb-database localhost -influxdb-user sensors -influxdb-user-password=S3CRET
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

> Note that compilation is not part of the deployment. This has to happen in a previous CI step.

## Troubleshooting

* Since the sensors bot is managed by systemd, the [usual ways](https://wiki.archlinux.org/index.php/Systemd#Troubleshooting) to inspect can be used.

    Examples:

    ```bash
    $ systemctl status sensors.{service,timer} # get an overview
    $ journalctl --unit sensors.service -f # tail the logs
    ```

# TODO

* Inject the version into the go binary [at build time](https://stackoverflow.com/a/11355611/3212907)
* Print the application version number at startup
