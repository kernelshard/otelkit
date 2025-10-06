# Traced HTTP Client Example


```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kernelshard/otelkit"
)

func main() {
	// Initialize tracing
	ctx := context.Background()
	shutdown, err := otelkit.SetupTracing(ctx, "traced-http-client-example")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	// Create a tracer
	tracer := otelkit.New("http-client-example")

	// Create a traced HTTP client
	client := otelkit.NewTracedHTTPClient(&http.Client{
		Timeout: 10 * time.Second,
	}, tracer, "jsonplaceholder-api")

	// Make a GET request
	ctx, span := tracer.Start(ctx, "example-http-request")
	defer span.End()

	resp, err := client.Get(ctx, "https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		otelkit.RecordError(span, err)
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Make a POST request
	postData := []byte(`{"title": "foo", "body": "bar", "userId": 1}`)
	resp2, err := client.Post(ctx, "https://jsonplaceholder.typicode.com/posts", "application/json", postData)
	if err != nil {
		otelkit.RecordError(span, err)
		log.Printf("POST request failed: %v", err)
	} else {
		defer resp2.Body.Close()
		fmt.Printf("POST Response Status: %s\n", resp2.Status)
	}

	fmt.Println("HTTP client tracing example completed. Check your tracing backend for spans.")
}
```

How to trace outbound HTTP requests with otelkit.