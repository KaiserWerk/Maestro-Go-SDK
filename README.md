# Maestro Go SDK
The Golang SDK for the [Maestro](https://github.com/KaiserWerk/Maestro) Service Discovery API.

## Usage

First, import the package:

``import maestro "github.com/KaiserWerk/Maestro-Go-SDK"``

Then, create a Client:
```golang
maestroUrl := "http://some-address.com" // no need for trailing slash
authToken := "some secret token"
id := "useful-service-idenitifier"
client := maestro.New(maestroUrl, authToken, id, nil) // some configuration is optional
```

Now, you can register your app:
```golang
// registers the running app with the Maestro service using the given address
err := client.Register("http://app-url")  
```

Similarly, you can deregister easily:

```golang
// deregisters the running app with the Maestro service
err := client.Deregister()  
```

Of course, querying the Maestro service for a specific service identifier is just as simple
(error omitted):
```golang
q, _ := client.Query("some-other-service-id")
fmt.Printf("Found address: %s\n", q.Address)
```

In order to signal the Maestro service that your registered app is alive and well, you should
send out health pings:
```golang
// call the cancel func when you want to stop the goroutine
// or create a context with timeout/deadline
ctx, cancel := context.WithCancel(context.Background())
// A 30 second interval is a solid default
go client.StartPing(ctx, 30*time.Second)
```


