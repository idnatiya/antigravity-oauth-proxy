//go:build !js || !wasm

package server

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
)

type callbackListener struct {
	mu     sync.Mutex
	server *http.Server
	ln     net.Listener
}

func (l *callbackListener) start(a *GoogleAuth) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.stopLocked()

	ln, err := net.Listen("tcp", "127.0.0.1:51121")
	if err != nil {
		logger.Get().Warn().Err(err).Msg("Could not start local OAuth callback listener on 127.0.0.1:51121 (manual code paste fallback available)")
		return
	}
	l.ln = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth-callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Missing authorization code.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := a.Complete(ctx, r.URL.String())
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;background:#121316;color:#ef4444;text-align:center;padding:3rem;"><h2>Authorization Failed</h2><p>%s</p></body></html>`, html.EscapeString(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Authorization Successful</title></head><body style="font-family:sans-serif;background:#121316;color:#fff;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;"><div style="text-align:center;padding:2.5rem;border:1px solid #27272a;border-radius:1rem;background:#18181b;max-width:420px;box-shadow:0 10px 25px rgba(0,0,0,0.5);"><div style="font-size:2.5rem;margin-bottom:0.75rem;">🎉</div><h2 style="color:#22c55e;margin-top:0;margin-bottom:0.5rem;font-size:1.25rem;">Account Connected!</h2><p style="color:#a1a1aa;font-size:0.875rem;line-height:1.5;">Your Google account is now ready in Antigravity Proxy. This window will close automatically.</p><script>setTimeout(function(){ window.close(); }, 1200);</script></div></body></html>`)

		go func() {
			time.Sleep(1 * time.Second)
			l.stop()
		}()
	})

	l.server = &http.Server{Handler: mux}
	go func() {
		if err := l.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Get().Debug().Err(err).Msg("OAuth callback server stopped")
		}
	}()
	logger.Get().Info().Msg("Started OAuth callback listener on http://127.0.0.1:51121/oauth-callback")
}

func (l *callbackListener) stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stopLocked()
}

func (l *callbackListener) stopLocked() {
	if l.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = l.server.Shutdown(ctx)
		l.server = nil
	}
	if l.ln != nil {
		_ = l.ln.Close()
		l.ln = nil
	}
}
