module go.elastic.co/apm/module/apmprometheus

require (
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/prometheus/client_model v0.2.0
	github.com/stretchr/testify v1.4.0
	go.elastic.co/apm v1.6.0
)

replace go.elastic.co/apm => ../..

go 1.13
