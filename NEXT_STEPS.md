# Next Steps

- App: IDP
- App: user context
- No CSRF protection for state-changing browser requests. This matters particularly if authentication later uses cookies
- Audit trail
- Add graceful shutdown: The server calls ListenAndServe directly and panics on unexpected termination in main.go:97-99. A production service should handle SIGTERM, stop accepting traffic, drain active requests and SSE streams, close the database pool, and exit within a bounded timeout.