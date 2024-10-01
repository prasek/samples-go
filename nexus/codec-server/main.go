package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/temporalio/samples-go/nexus/options"
	"go.temporal.io/sdk/converter"
)

func main() {
	codecCorsURL := "https://cloud.temporal.io"
	codecPort := 8088
	if codecCorsURL != "" {
		log.Printf("Codec Server will allow requests from Temporal Web UI at: %s\n", codecCorsURL)

		if strings.HasSuffix(codecCorsURL, "/") {
			// In my experience, a slash character at the end of the URL will
			// result in a "Codec server could not connect" in the Web UI and
			// the cause will not be obvious. I don't want to strip it off, in
			// case there really is a valid reason to have one, but warning the
			// user could help them to more quickly spot the problem otherwise.
			log.Println("Warning: Temporal Web UI base URL ends with '/'")
		}
	}

	log.Printf("Starting Codec Server on port %d\n", codecPort)
	err := RunCodecServer(codecPort, codecCorsURL)
	if err != nil {
		log.Fatalf("Unable to start Codec Server on port %d: %s\n", codecPort, err)
	}
}

func newCORSHTTPHandler(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RunCodecServer launches the Codec Server on the specified port, enabling
// CORS for the Temporal Web UI at the specified URL
func RunCodecServer(port int, url string) error {
	// The EncryptionKeyID attribute is omitted when creating the Codec
	// instance below because the Codec Server only decrypts. It locates
	// the encryption key ID from the payload's metadata.
	handler := converter.NewPayloadCodecHTTPHandler(&options.Codec{})

	if url != "" {
		handler = newCORSHTTPHandler(url, handler)
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		_ = srv.Close()
	case err := <-errCh:
		return err
	}

	return nil
}
