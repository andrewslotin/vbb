# Go client for Berlin/Brandenburg transport API

A Go client for the [VBB API](https://v5.vbb.transport.rest/api.html)

## Usage

Add to your `go.mod` file:

```bash
go get github.com/andrewslotin/vbb
```

Initialize the API client:

```go
c := vbb.New(vbb.BaseURL, nil)
```
