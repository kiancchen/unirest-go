# Unirest-Go

A Golang http client with [unirest-java](http://kong.github.io/unirest-java/) styled interfaces. It encapsulates the native `net/http` library to provide simplified interfaces.

# Example

```go
response, err := New().SetURL("https://www.google.com").
									 AppendPath("/search").
									 AddQuery("q","unirest").Send().AsString() // or AsBytes()
```

## Add other request parameters

- AddHeader
- AddFormField
- AddFile
- SetJSONBody
- SetRawBody
- SetBasicAuth

## HTTP method

To make it simplified, it automatically chooses to use GET or POST method, depending whether you send parameters in the body. If you want to do it manually, you can call methods `Get` or `Post`. More request methods will be provided in the future.

# Get native http request instance

Use `ParseRequest` method to produce a native http request after adding parameters.

# Turn off auto clone

To make the client can be reused, it will clone the client instance every time you change it. You can use `AutoClone(false)` to turn off it or set true to turn on it. And you can use `Clone` to make clone manually.

