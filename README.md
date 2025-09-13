# bobok
A lightweight and concurrent Pub/Sub messaging library for Go. 📨

<p align="center">
  <img width="320px" src="https://github.com/rramirezacosta/bobok/blob/main/bobok.webp?raw=true" alt="bobok"/>
</p>

**Features**:

- **Lightweight**: Minimal overhead and dependencies.
- **Concurrent**: Built with goroutines in mind for high-performance applications.
- **Simple API**: Get started with intuitive and easy-to-use methods.


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
ch, unsubscribe, _ := bobok.Subscribe("topic_name")
defer unsubscribe()

// Listen for messages
rawMsg := <-ch
fmt.Println("Received message:", rawMsg.(string))

```

Publish:
```go
// Publish a message to a topic_name
bobok.Publish("topic_name", "Hello, World!")
```

# Why "Bobok"
<dl><dd>
When a terrible drought struck the Yaqui lands, the sparrow and the swallow carried pleas to the rain god, Yuku, but no answer came. All hope seemed lost until they called upon Bobok, a wise toad who knew delivery was everything. On borrowed bat wings, he flew to the heavens and confronted the god. With blunt conviction, he demanded Yuku do his duty, proving to be the only messenger capable of the task.
</dd></dl>
