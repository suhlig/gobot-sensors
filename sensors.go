package main

import (
	"flag"
	"fmt"
	"os"

	influx "github.com/influxdata/influxdb/client/v2"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

var influxURL = flag.String("influxdb-url", "http://localhost:8086", "URL to the InfluxDB where samples are sent to")
var influxDatabase = flag.String("influxdb-database", "", "InfluxDB database name where samples are written to")
var influxUsername = flag.String("influxdb-user", "", "InfluxDB user name that can write samples to the given database.")
var influxPassword = flag.String("influxdb-password", "", "Password for influxdb-user.")
var i2cBus = flag.String("i2c-bus", "", "I²C bus to use. By default the first I²C bus found is used.")
var i2cAddress = flag.Uint("i2c-address", 0x77, "I2C address of the BME280 device")

const hecto = 100

func newPoint(key string, value float64, tags map[string]string) (*influx.Point, error) {
	fields := map[string]interface{}{
		"value": value,
	}

	return influx.NewPoint(key, tags, fields)
}

func publish(influxClient influx.Client, measurements physic.Env) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  *influxDatabase,
		Precision: "s",
	})

	if err != nil {
		return err
	}

	tags := map[string]string{}

	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	tags["host"] = hostname

	if err != nil {
		return err
	}

	humidity, err := newPoint("humidity", float64(measurements.Humidity)/float64(physic.PercentRH), tags)

	if err != nil {
		return err
	}

	bp.AddPoint(humidity)

	temperature, err := newPoint("temperature", measurements.Temperature.Celsius(), tags)

	if err != nil {
		return err
	}

	bp.AddPoint(temperature)

	pressure, err := newPoint("pressure", float64(measurements.Pressure)/float64(physic.Pascal)/hecto, tags)

	if err != nil {
		return err
	}

	bp.AddPoint(pressure)

	return influxClient.Write(bp)
}

func main() {
	flag.Parse()

	if *influxDatabase == "" {
		fmt.Fprintf(os.Stderr, "Error: missing mandatory database name.\n")
		os.Exit(1)
	}

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     *influxURL,
		Username: *influxUsername,
		Password: *influxPassword,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to InfluxDB at %v: %v\n", *influxURL, err)
		os.Exit(1)
	}

	s := bmxx80.O4x
	opts := bmxx80.Opts{Temperature: s, Pressure: s, Humidity: s}

	if _, err := host.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize I2C host: %s\n", err)
		os.Exit(1)
	}

	var dev *bmxx80.Dev
	i, err := i2creg.Open(*i2cBus)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open I2C bus: %s\n", err)
		os.Exit(1)
	}
	defer i.Close()

	if dev, err = bmxx80.NewI2C(i, uint16(*i2cAddress), &opts); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create I2C device: %s\n", err)
		os.Exit(1)
	}

	measurements := physic.Env{}
	if err := dev.Sense(&measurements); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read sensor: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Humidity: %v\n", measurements.Humidity)
	fmt.Printf("Temperature: %v\n", measurements.Temperature)
	fmt.Printf("Pressure: %v\n", measurements.Pressure)

	err = publish(influxClient, measurements)

	if err != nil {
		fmt.Println("Error publishing sensor data to InfluxDB: ", err)
		os.Exit(1)
	}
}
