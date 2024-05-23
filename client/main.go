package main

import (
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
)

func main() {
	app := fiber.New()

	app.Get("/api", api)
	app.Get("/api2", api2)

	app.Listen(":8001")
}

func init() {
	// 20 req / 10sec have 50% error rate circuit will open (default)
	// when not have error circuit will close
	// hystrix.DefaultTimeout = 500 // apply to all api
	hystrix.ConfigureCommand("api", hystrix.CommandConfig{
		Timeout:                500, // default 1000
		RequestVolumeThreshold: 1,   // default 20
		ErrorPercentThreshold:  100, // default 50% threshold
		SleepWindow:            15000,
		// when circuit open will sleep 15 sec, then half open
		// when half open and error, circuit will open again
	})

	hystrix.ConfigureCommand("api2", hystrix.CommandConfig{
		Timeout:                500,
		RequestVolumeThreshold: 10,
		ErrorPercentThreshold:  50,
		SleepWindow:            10000,
	})

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	// can start 2 server in 1 thread, need to use go routine
	go http.ListenAndServe(":8002", hystrixStreamHandler)
	// use hystrix dashboard to read stream data
	// http://host.docker.internal:8002 use in docker
}

func api(c *fiber.Ctx) error {
	output := make(chan string, 1)
	hystrix.Go("api", func() error {
		res, err := http.Get("http://localhost:8000/api")
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		msg := string(data)
		fmt.Println(msg)

		output <- msg
		return nil
	}, func(err error) error { // if failure hystrix will call fallback
		fmt.Println(err)
		return nil
	})

	out := <-output
	return c.SendString(out)
}

func api2(c *fiber.Ctx) error {
	output := make(chan string, 1)
	hystrix.Go("api2", func() error {
		res, err := http.Get("http://localhost:8000/api")
		if err != nil {
			return err
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		msg := string(data)
		fmt.Println(msg)

		output <- msg
		return nil
	}, func(err error) error { // if failure hystrix will call fallback
		fmt.Println(err)
		return nil
	})

	out := <-output
	return c.SendString(out)
}
