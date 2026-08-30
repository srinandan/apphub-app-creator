// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

// ServerCmd starts an HTTP server
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start an HTTP server",
	Long:  "Starts a HTTP server for the App Hub Creator.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		// Deadline the server waits for in-flight requests to drain on shutdown.
		const shutdownTimeout = 15 * time.Second

		logger := slog.Default()

		r := mux.NewRouter()
		r.HandleFunc("/healthz", healthHandler).Methods("GET")
		r.HandleFunc("/generate", generateHandler).Methods("POST")
		// registerUI owns the root path: a liveness endpoint by default, or the
		// embedded web UI when the binary is built with the "embedui" tag. It is
		// registered last so its catch-all never shadows the API routes above.
		registerUI(r)

		// Create a new CORS handler with specific options.
		corsHandler := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "ApiKey", "User"},
			Debug:            false,
		})

		// Get the port from the flag
		port := GetStringParam(cmd.Flag("port"))
		addr := ":" + port

		srv := &http.Server{
			Addr:         addr,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      corsHandler.Handler(r),
		}

		// Start the server
		logger.Info("Starting server...", "address", addr)
		go func() {
			// ListenAndServe always returns a non-nil error; ErrServerClosed is
			// the expected result of a graceful Shutdown and must not be fatal.
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("Error starting server", "error", err)
				os.Exit(1)
			}
		}()

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)

		// Block until we receive our signal.
		<-c

		// Create a deadline to wait for.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline.
		_ = srv.Shutdown(ctx)

		logger.Info("Shutting down server")
		return nil
	},
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func init() {
	var port string

	ServerCmd.Flags().StringVarP(&port, "port", "p",
		"8080", "Port to listen on")
}
