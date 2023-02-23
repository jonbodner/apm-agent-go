module go.elastic.co/apm/module/apmgopg

require (
	github.com/go-pg/pg v8.0.4+incompatible
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/stretchr/testify v1.3.0
	go.elastic.co/apm v1.6.0
	go.elastic.co/apm/module/apmsql v1.6.0
	golang.org/x/text v0.3.8 // indirect
	mellium.im/sasl v0.2.1 // indirect
)

replace go.elastic.co/apm => ../..

replace go.elastic.co/apm/module/apmsql => ../apmsql

go 1.13
