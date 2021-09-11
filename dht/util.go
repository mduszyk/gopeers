package dht

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

type Payload = interface{}
type ProcessResult struct {
    value Payload
    err error
}
type ProcessFn = func(payload Payload) (Payload, error)

func Pool(n int, f ProcessFn) (chan Payload, chan ProcessResult) {
    input := make(chan Payload, n)
    output := make(chan ProcessResult, n)
    for i := 0; i < n; i++ {
       go func() {
           for {
               payload, more := <-input
               if more {
                   value, err := f(payload)
                   output <- ProcessResult{value, err}
               } else {
                   return
               }
           }
       }()
    }
    return input, output
}
