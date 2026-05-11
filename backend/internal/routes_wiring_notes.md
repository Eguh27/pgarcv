Route wiring notes

Public routes to add in cmd/server/main.go
- api.GET("/videos/suggest", videoSuggestH.Suggest)
- api.GET("/videos/:id/download", videoDownloadH.Download)

Admin routes already have upload/static.

