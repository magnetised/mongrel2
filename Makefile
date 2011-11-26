include $(GOROOT)/src/Make.inc

TARG=mongrel2

GOFILES=\
	http_handler.go\
	raw.go\
	spec.go\
	json_handler.go

include $(GOROOT)/src/Make.pkg