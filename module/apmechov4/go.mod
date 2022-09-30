module go.elastic.co/apm/module/apmechov4

require (
	github.com/labstack/echo/v4 v4.9.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.7.0
	go.elastic.co/apm v1.6.0
	go.elastic.co/apm/module/apmhttp v1.6.0
)

replace go.elastic.co/apm => ../..

replace go.elastic.co/apm/module/apmhttp => ../apmhttp

go 1.13
