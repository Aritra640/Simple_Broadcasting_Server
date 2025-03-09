package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()
	e.Use(middleware.CORS())

	b := newBroadcast()
	go b.run()

	e.Any("/join_broadcast", b.WShandle)

	//Create a channel to signal to listen for OS interupts signals (ctrl + c)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		close(b.stop)
		e.Shutdown(context.Background())
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
