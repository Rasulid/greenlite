package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)


func (app *application) server() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}


	shutDownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit


		app.logger.PrintInfo("shutdown server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutDownError <- err
		}
		
		app.logger.PrintInfo("Complating background tasks", map[string]string{
			"adr": srv.Addr,
		})


		app.wg.Wait()
		shutDownError <- nil
		
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": app.Config.environment,
		})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed){
		return err
	}

	err = <- shutDownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}


func pri(){
	print("ss")
}