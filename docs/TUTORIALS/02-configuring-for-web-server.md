# Tutorial: Caching in a Web Server

This tutorial demonstrates how to use Metis to cache responses in a simple Go web server, a common and effective use case for improving application performance.

## Prerequisites

- Completed the [Basic Usage Tutorial](./01-basic-usage.md).
- Basic understanding of Go's `net/http` package.

## The Scenario

We will build a simple API endpoint `/users/{id}` that simulates a slow database query. We will then use Metis to cache the results of this query to make subsequent requests for the same user instantaneous.

## Step 1: Set Up the Project

Let's create a new `main.go` file for our web server.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/agilira/metis"
)

// User represents our data model.
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// A global cache instance for our application.
var cache *metis.Cache

func main() {
    // Configure Metis for a web server environment.
    // WTinyLFU is a great default for varied access patterns.
    config := metis.CacheConfig{
        CacheSize:      1000,
        EvictionPolicy: "wtinylfu",
        TTL:            10 * time.Minute,
    }
    cache = metis.NewWithConfig(config)
    defer cache.Close()

    // Set up our HTTP routes.
    http.HandleFunc("/users/", userHandler)

    fmt.Println("Server starting on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// userHandler handles requests to /users/{id}.
func userHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/users/"):]
    cacheKey := "user:" + id

    // 1. Check the cache first.
    if cachedUser, found := cache.Get(cacheKey); found {
        log.Printf("Cache HIT for user %s", id)
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Cache-Status", "HIT")
        json.NewEncoder(w).Encode(cachedUser)
        return
    }

    // 2. If not in cache, perform the "slow" operation.
    log.Printf("Cache MISS for user %s", id)
    user, err := fetchUserFromDatabase(id)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // 3. Store the result in the cache for next time.
    cache.Set(cacheKey, user)

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Cache-Status", "MISS")
    json.NewEncoder(w).Encode(user)
}

// fetchUserFromDatabase simulates a slow database query.
func fetchUserFromDatabase(id string) (User, error) {
    log.Printf("Performing slow database query for user %s...", id)
    time.Sleep(2 * time.Second) // Simulate latency
    
    // In a real app, you'd query a database.
    // Here, we'll just return a dummy user.
    if id == "1" {
        return User{ID: "1", Name: "Alice"}, nil
    }
    if id == "2" {
        return User{ID: "2", Name: "Bob"}, nil
    }
    return User{}, fmt.Errorf("user not found")
}
```

## Step 2: Run the Server

Start the web server.

```bash
go run main.go
```

## Step 3: Test the Endpoint

Open a new terminal and use `curl` to test the API.

**First Request:**
Make a request for user `1`. This will be a cache "MISS".

```bash
time curl http://localhost:8080/users/1
```

- **Server Logs**: You will see `Cache MISS` and `Performing slow database query`.
- **`curl` Output**: The request will take about 2 seconds.

**Second Request:**
Immediately make the same request again. This will be a cache "HIT".

```bash
time curl http://localhost:8080/users/1
```

- **Server Logs**: You will see `Cache HIT`.
- **`curl` Output**: The request will be almost instantaneous.

## Conclusion

You have successfully integrated Metis into a web server to dramatically improve response times for repeated requests. This pattern is fundamental to building high-performance, scalable applications.
