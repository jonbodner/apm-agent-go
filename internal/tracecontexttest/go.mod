module tracecontexttest

require go.elastic.co/apm/module/apmhttp v1.6.0

replace go.elastic.co/apm => ../..

replace go.elastic.co/apm/module/apmhttp => ../../module/apmhttp

go 1.13
