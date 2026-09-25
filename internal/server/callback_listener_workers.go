//go:build js && wasm

package server

type callbackListener struct{}

func (l *callbackListener) start(a *GoogleAuth) {}
func (l *callbackListener) stop()               {}
