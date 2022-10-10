module github.com/Reid00/go-raft

go 1.19

require (
	github.com/golang/protobuf v1.5.2
)

require google.golang.org/protobuf v1.26.0 // indirect

// Bad imports are sometimes causing attempts to pull that code.
// This makes the error more explicit.
replace github.com/Reid00/go-raft/api => ./api
