# Tutorial: Custom Serialization

Metis uses Go's built-in `gob` package for serialization, which is necessary for features like compression. By default, `gob` can only encode standard Go types. If you want to cache instances of your own custom structs, you must first "register" them with the `gob` package.

This tutorial explains why and how to register your custom types.

## Prerequisites

- Completed the [Basic Usage Tutorial](./01-basic-usage.md).
- You are trying to cache a custom struct and have enabled `Compress: true`.

## The Problem

Consider the following code where we try to cache a `CustomData` struct with compression enabled.

```go
package main

import (
    "fmt"
    "github.com/agilira/metis"
)

// A custom struct we want to cache.
type CustomData struct {
    Name  string
    Value int
}

func main() {
    // This will PANIC!
    cache := metis.NewWithConfig(metis.CacheConfig{
        Compress: true, // Compression requires serialization
    })
    defer cache.Close()

    data := CustomData{Name: "example", Value: 42}

    // The Set call will fail because gob doesn't know about CustomData.
    cache.Set("key", data)

    fmt.Println("This line will not be reached.")
}
```

If you run this code, it will panic with an error similar to this:

```
panic: gob: type not registered for interface: main.CustomData
```

This happens because when Metis tries to serialize the `CustomData` struct before compressing it, the `gob` package has no knowledge of this type and refuses to proceed.

## The Solution: `gob.Register()`

To fix this, you must explicitly register your custom type with the `gob` package at the start of your application. The `init` function is the perfect place for this, as it's guaranteed to run before `main`.

### Step 1: Register the Type

Modify your `main.go` to include an `init` function that calls `gob.Register`.

```go
package main

import (
    "encoding/gob" // 1. Import the gob package
    "fmt"
    "github.com/agilira/metis"
)

// A custom struct we want to cache.
type CustomData struct {
    Name  string
    Value int
}

// 2. Register the type in an init function.
func init() {
    gob.Register(CustomData{})
}

func main() {
    // Now this will work perfectly.
    cache := metis.NewWithConfig(metis.CacheConfig{
        Compress: true,
    })
    defer cache.Close()

    data := CustomData{Name: "example", Value: 42}

    // The Set call will now succeed.
    cache.Set("key", data)

    fmt.Println("Successfully cached custom data!")

    // Retrieve and type-assert the data.
    if retrieved, found := cache.Get("key"); found {
        if castedData, ok := retrieved.(CustomData); ok {
            fmt.Printf("Retrieved: %+v\n", castedData)
        }
    }
}
```

### Why `gob.Register(CustomData{})`?

-   `gob.Register()` tells the `gob` encoder/decoder about the type `CustomData`.
-   We pass an empty instance `CustomData{}` to provide the type information.
-   By placing this in `init()`, we ensure that the type is registered before any cache operations can occur.

## When Do I Need to Register Types?

You need to call `gob.Register()` for your custom types if you meet **both** of these conditions:

1.  You are using a Metis feature that requires serialization. Currently, this is **Compression (`Compress: true`)**.
2.  The value you are storing in the cache is a **custom struct** or a type that contains a custom struct.

If you are only storing basic types like `string`, `[]byte`, `int`, etc., you do not need to register them, even with compression enabled.

## Registering Multiple Types

If you have multiple custom types that you plan to cache, simply register them all in the `init` function.

```go
type User struct{ /* ... */ }
type Order struct{ /* ... */ }
type Product struct{ /* ... */ }

func init() {
    gob.Register(User{})
    gob.Register(Order{})
    gob.Register(Product{})
}
```
