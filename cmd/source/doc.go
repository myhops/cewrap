/*
The source command implements a proxy server that emits cloud events when
it successfully handled a method that changes the data in the downstream service.

It logs to standard error, currently in structured text format and with Info log level.

	Usage:

		source \
			-downstream http://service.example.com/downstream/service \
			-sink http://broker.mynamespace.svc.cluster.local \
			-source http://service.example.com/crm \
			-port 8080 \
			-dataschema http://schema.example.com \
			-type-prefix com.example.service.crm \
			-path-prefix /downstream/service

	Options:
		-downstream
		-sink
		-source
		-port
		-dataschema
		-type-prefix
		-path-prefix

And so on
*/
package main