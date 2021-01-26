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
var i2cBus = flag.Int("i2c-bus", 0, "I²C bus to use for all sensors")
var bme280Address = flag.Int("bme280-address", 0, "I2C address of the BME280 device")
var tsl2561Address = flag.Int("tsl2561-address", 0, "I2C address of the TSL2561 device")

func publish(influxClient influx.Client, key string, value float64) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  *influxDatabase,
		Precision: "s",
	})

	if err != nil {
		return err
	}

	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	pt, err := influx.NewPoint(
		key,
		map[string]string{"host": hostname},
		map[string]interface{}{"value": value},
	)

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

	var luxSensor *i2c.TSL2561Driver
	var bme280 *i2c.BME280Driver

	if *tsl2561Address != 0 {
		luxSensor = i2c.NewTSL2561Driver(raspi, i2c.WithBus(*i2cBus), i2c.WithAddress(*tsl2561Address), i2c.WithTSL2561Gain16X)
	}

	if *bme280Address != 0 {
		bme280 = i2c.NewBME280Driver(raspi, i2c.WithBus(*i2cBus), i2c.WithAddress(*bme280Address))
	}

	influxPassword, found := os.LookupEnv("INFLUXDB_PASSWORD")

	if !found {
		fmt.Fprintln(os.Stderr, "Could not find the INFLUXDB_PASSWORD environment variable")
		os.Exit(1)
	}

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     *influxURL,
		Username: *influxUsername,
		Password: influxPassword,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to InfluxDB at %v: %v\n", *influxURL, err)
		os.Exit(1)
	}

	bot := gobot.NewRobot("sensors-bot", []gobot.Connection{raspi}, func() {
		gobot.Every(time.Minute, func() {
			if luxSensor != nil {
				broadband, ir, err := luxSensor.GetLuminocity()

				if err != nil {
					fmt.Println("Error reading luminocity: ", err)
				} else {
					light := luxSensor.CalculateLux(broadband, ir)

					if light > 10000 {
						fmt.Printf("Warning: Ignoring value of %v lux; this seems to be an outlier.\n", light)
					} else {
						fmt.Printf("Light: %v lux\n", light)
						err = publish(influxClient, "light", float64(light))

						if err != nil {
							fmt.Println("Error publishing light to InfluxDB: ", err)
						}
					}
				}
			}

			if bme280 != nil {
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
			}
		})
	},
	)

	if luxSensor != nil {
		bot.AddDevice(luxSensor)
	}

	if bme280 != nil {
		bot.AddDevice(bme280)
	}

	bot.Start()
}
