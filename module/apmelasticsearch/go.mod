module go.elastic.co/apm/module/apmelasticsearch

require (
	github.com/stretchr/testify v1.3.0
	go.elastic.co/apm v1.6.0
	go.elastic.co/apm/module/apmhttp v1.6.0
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
)

replace go.elastic.co/apm => ../..

replace go.elastic.co/apm/module/apmhttp => ../apmhttp

go 1.13
