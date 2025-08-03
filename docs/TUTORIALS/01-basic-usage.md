# Tutorial: Basic Usage

This tutorial will guide you through the process of setting up Metis in a new Go project and performing basic cache operations.

## Prerequisites

- Go 1.18 or later installed.
- A working Go environment.

## Step 1: Initialize a New Project

First, let's create a new directory for our project and initialize a Go module.

```bash
mkdir metis-example
cd metis-example
go mod init metis-example
```

## Step 2: Install Metis

Next, add Metis as a dependency to your project.

```bash
go get github.com/agilira/metis
```

## Step 3: Create a `main.go` File

Create a file named `main.go` in your project directory. This is where we will write our code.

```go
package main

import (
    "fmt"
    "github.com/agilira/metis"
    "time"
)

func main() {
    // For this example, we will configure the cache programmatically.
    config := metis.CacheConfig{
        CacheSize:      100, // Keep it small for the example
        EvictionPolicy: "lru",
        TTL:            5 * time.Second,
    }

    // Create a new cache instance.
    cache := metis.NewWithConfig(config)
    fmt.Println("Metis cache initialized.")

    // Always remember to close the cache to free up resources.
    defer cache.Close()

    // --- Basic Operations ---

    // 1. Set a value
    fmt.Println("\nSetting key 'user:1' with value 'Alice'...")
    cache.Set("user:1", "Alice")

    // 2. Get a value
    value, found := cache.Get("user:1")
    if found {
        fmt.Printf("Found 'user:1': %s\n", value.(string))
    } else {
        fmt.Println("'user:1' not found.")
    }

    // 3. Get a non-existent value
    _, found = cache.Get("user:2")
    if !found {
        fmt.Println("As expected, 'user:2' was not found.")
    }

    // 4. Demonstrate TTL expiration
    fmt.Println("\nWaiting for 6 seconds to let the TTL expire...")
    time.Sleep(6 * time.Second)

    _, found = cache.Get("user:1")
    if !found {
        fmt.Println("'user:1' has been evicted due to TTL expiration.")
    }

    // 5. Check cache stats
    stats := cache.Stats()
    fmt.Printf("\nFinal Cache Stats: %+v\n", stats)
}
```

## Step 4: Run the Code

Execute the program from your terminal.

```bash
go run main.go
```

### Expected Output

You should see an output similar to this:

```
Metis cache initialized.

Setting key 'user:1' with value 'Alice'...
Found 'user:1': Alice
As expected, 'user:2' was not found.

Waiting for 6 seconds to let the TTL expire...
'user:1' has been evicted due to TTL expiration.

Final Cache Stats: {Size:0 Hits:1 Misses:3 HitRate:25.00}
```

## Conclusion

Congratulations! You have successfully set up Metis, stored and retrieved data, and observed its TTL-based eviction in action.

In the next tutorial, we will explore how to configure Metis for a web server environment.
