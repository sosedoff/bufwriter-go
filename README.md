# bufwriter

A simple buffered writer implementation

## Usage

Basic example:

```go
package main

import(
  "github.com/sosedoff/bufwriter-go"
)

func main() {
  // Buffer up to 1mb before writing to STDOUT
  writer := bufwriter.New(1024 * 1024, os.Stdout)

  for n := 0; n < 100; n++ {
    // Write calls will buffer contents into memory until full, then call Flush
    n, err := writer.Write(...)
  }

  // Flush all buffered content
  writer.Flush()
}
```

With periodic flusher:

```go
writer := bufwriter.New(1024 * 1024, os.Stdout)

// Flush buffer contents periodically
go writer.StartFlusher(time.Second, func(err error) {
  // handle error
})
defer writer.Stop()

// Write stuff
n, err := writer.Write(...)
```
