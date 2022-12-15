// Package coinbase is a Coinbase SDK but with a bit more of the consumer needs considered.
// This may be better designed as a pure SDK, but this would have involved another
// package + abstraction which seemed excessive.
package coinbase

//go:generate moq -out moq_test.go . Conn Dialer
