Mongrel2 Binding For Go
=======================

This package provides you with the code needed to write handlers for mongrel2
in go.  

This package assumes that you have already installed 
[mongrel2](http://mongrel2.org) version 1.7.5+,
[the 0mq library](http://www.zeromq.org/) version 2.1+,
and the 
[0mq support for go](https://github.com/alecthomas/gozmq). 

To run gozmq with
later versions of go (such as weekly of 9 November 2011), requires a 
[patch](https://gist.github.com/1362154)
for gozmq.

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

In Eclipse
----------

The project files are designed to be used with
[goclipse](http://code.google.com/p/goclipse/) which may help explain the
unusual directory layout.  If you load this project into goclipse, you can just 
use the "Run As Go Program" menu option to run either of the two supplied 
handlers, `mongrel_raw` or `workers`.  Because the package mongrel2 is in the 
src/pkg directory, you should not need to build anything and can ignore
the supplied Makefiles.

### Troubleshouting in Eclipse

The goclipse support is quite primitive.  Here are two common things that 
"make it happy."

* Use the menu option "Project > Clean" liberally.  Goclipse seems to get 
confused easily, especially if you are moving files around.  If it says there 
is a mistake on a line, especially if it is an error about an uknown type or 
name that it clearly _should_ know about, this should be your first response.

* Make sure that check the "Problems" view in Eclipse to see if there are any clues 
there.  Often, it will complain that it cannot open a directory like
"bin/darwin\_amd64" or "pkg/darwin\_amd64".  Goclipse _should_ be smart enough to
create these for you, since they are just the place it puts its compiled
binaries, but it doesn't.  You can go to the top level of the project in the
workspace (the one with "src" as a child) and create the directories it
needs, such as "bin/darwin\_amd64" and "pkg/darwin\_amd64" on a mac. If you need
to know your directory name like "darwin_amd64" it is just the value of your
`GOOS` and `GOARCH` environment variables.