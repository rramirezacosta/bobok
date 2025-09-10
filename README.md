# bobok
A lightweight and concurrent Pub/Sub messaging library for Go. 📨

<p align="center">
  <img width="320px" src="https://github.com/rramirezacosta/bobok/blob/main/bobok.webp?raw=true" alt="bobok"/>
</p>

# Installation
```bash
### Install the package
go get github.com/rramirezacosta/bobok
```

# Usage
Import:
```go 
// Import
import "github.com/rramirezacosta/bobok"
```

Subscribe and Listen:
```go
// Suscribe to a topic_name
ch, done, cleanup := bobok.Subscribe("topic_name")
defer cleanup()

// Listen for messages
select {
case raw := <-ch:
    msg := raw.(string)
    fmt.Println("Received message:", msg)
case <-done:
    fmt.Println("Subscription closed")
    break
}

```

Publish:
```go
// Publish a message to a topic_name
bobok.Publish("topic_name", "Hello, World!")
```
