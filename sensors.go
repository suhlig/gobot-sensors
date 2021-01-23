package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"

	influx "github.com/influxdata/influxdb/client/v2"
)

var influxURL = flag.String("influxdb-url", "http://localhost:8086", "URL to the InfluxDB where samples are sent to")
var influxDatabase = flag.String("influxdb-database", "", "InfluxDB database name where samples are written to")
var influxUsername = flag.String("influxdb-user", "", "InfluxDB user name that can write samples to the given database.")

func publish(influxClient influx.Client, key string, value float64) error {
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

	fields := map[string]interface{}{
		"value": value,
	}

	pt, err := influx.NewPoint(key, tags, fields)

	if err != nil {
		return err
	}

	bp.AddPoint(pt)

	return influxClient.Write(bp)
}

func main() {
	flag.Parse()

	if *influxDatabase == "" {
		fmt.Fprintf(os.Stderr, "Error: missing mandatory database name.\n")
		os.Exit(1)
	}

	raspi := raspi.NewAdaptor()
	bme280 := i2c.NewBME280Driver(raspi)

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     *influxURL,
		Username: *influxUsername,
		Password: os.Getenv("INFLUXDB_PASSWORD"),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to InfluxDB at %v: %v\n", *influxURL, err)
		os.Exit(1)
	}

	work := func() {
		gobot.Every(10*time.Second, func() {
			humidity, err := bme280.Humidity()

			if err != nil {
				fmt.Println("Error reading humidity: ", err)
				os.Exit(1)
			}
			fmt.Printf("Humidity: %v %%\n", humidity)
			err = publish(influxClient, "humidity", float64(humidity))

			if err != nil {
				fmt.Println("Error publishing humidity to InfluxDB: ", err)
				os.Exit(1)
			}

			temperature, err := bme280.Temperature()

			if err != nil {
				fmt.Println("Error reading temperature: ", err)
				os.Exit(1)
			}

			fmt.Printf("Temperature: %v °C\n", temperature)
			err = publish(influxClient, "temperature", float64(temperature))

			if err != nil {
				fmt.Println("Error publishing temperature to InfluxDB: ", err)
				os.Exit(1)
			}

			pressure, err := bme280.Pressure()

			if err != nil {
				fmt.Println("Error reading pressure: ", err)
				os.Exit(1)
			}

			hPa := pressure / 100 // Grafana wants hectopascal
			fmt.Printf("Pressure: %v hPa\n", hPa)
			err = publish(influxClient, "pressure", float64(hPa))

			if err != nil {
				fmt.Println("Error publishing pressure to InfluxDB: ", err)
				os.Exit(1)
			}
		})
	}

	robot := gobot.NewRobot("gobot",
		[]gobot.Connection{raspi},
		[]gobot.Device{bme280},
		work,
	)

	robot.Start()
}
