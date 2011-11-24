Mongrel2 Binding For Go
=======================

This package provides you with the code needed to write handlers for mongrel2
in go.  This package can be used separately from *Seven5*, but that would be
a mistake, no?

This package assumes that you have already installed 
[mongrel2](http://mongrel2.org) version 1.7.5+,
[the 0mq library](http://www.zeromq.org/) version 2.1+,
and the 
[0mq support for go](https://github.com/alecthomas/gozmq). 

Design
------
This is the mongrel2 layer.  Most *Seven5* developers can safely ignore this layer as it is pretty "raw" (like a _steack tartare_ kinda raw).

There are two important interfaces in this package, `M2HttpHandler` and `M2JsonHandler` that can respond to mongrel2 messages for (implementation of) HTTP requests and Json requests, respectively.  Each of these has an implementation of basic functionality called `M2HttpHandlerDefault` and `M2JsonHandlerDefault`, respectively.  These particular implementations know how to correctly `Bind()` to mongrel server--thus allocating resources like ZMQ sockets--and how to `Shutdown()` which does the reverse.  The two interfaces (and their implementations) know 
the particulars of the communication format that they should use with mongrel2.  For example the `M2HttpHandlerDefault` knows about `M2HttpRequest` and `M2HttpHandlerResponse`

There is an interface `M2RawHandler` and a corresponding simple implementation called `M2RawHandlerDefault` that should never be needed by developers.  It exists only to share code between the JS and HTTP code and thus avoid repetition.  Note that the implementations of
`M2HttpHandlerDefault` and `M2JsonHandlerDefault` include the implementation of `M2RawHandlerDefault`.  This is what you want.  If you had a composite that included the implementation of _both_ `M2HttpHandlerDefault` and `M2JsonHandlerDefault` it would need, and 
would have, four sockets for mongrel2 communication--since mongrel two considers these "different
handlers" from its point of view.  

Interfaces are structs prefixed by "M2" indicate that the entity deals with the raw mongrel2 layer.  This is to avoid confusion between, for example, the specific protocol that mongrel2 uses for talking with HTTP handlers and the HTTP protocol proper.

Install
-------

You must have your go 
[environment variables](http://golang.org/doc/install.html#environment), such as 
`GOROOT` set.

Go into the `src/pkg/mongrel2` and do `make install` to install the mongrel2
library into your go repository of packages.

Go into the `src/cmd/` directory and do `make` to build an example program. The
program expects that you have previously configured mongrel2 to expect a
hadler like this:

	handler_test = Handler(	send_spec='tcp://127.0.0.1:10070',
    	                   	send_ident='34f9ceee-cd52-4b7f-b197-88bf2f0ec378',
                  	     	recv_spec='tcp://127.0.0.1:10071',
							recv_ident='') 

You also need to map that handler to a path. Using the example configuration 
file that comes with mongrel2, this would be something like this:

    hosts = [
        Host(name="localhost", routes={
            '/tests/': Dir(base='tests/', index_file='index.html',default_ctype='text/plain')
	    '/handlertest': handler_test
        })
    ]

Running 
-------

Make sure mongrel2 is running.  We'll assume it's running on port 6767.

Running the `mongrel_raw` executable.  You should see

> waiting on a message from the mongrel2 server...

Fire a up a browser and point it at `http://localhost:6767/handlertest/foo`, 
assuming you are using the test configuration that comes with mongrel2 plus
the configuration info above.  You
should a message in the browser window plus some output on the command line
indicating that `mongrel_raw` responded to the request.

The `mongrel_raw` program only sends a single response to a single 
request. Thanks to 0mq, you can kill and restart handlers anytime while
leaving mongrel2 running.  You don't need to worry about startup order either..

If you want to see an example with several workers (goroutines) implementing the
handler, you
want to run the `mongrel_workers` handler.  Sadly, you have change the
`Makefile` to make it build `mongrel_workers` instead of `mongrel_raw`.  

If you run that executable, you can send multiple requests to the handler and
you will see in the browser results that there are different goroutines
responding, and in round-robin fashion.
ment variables.