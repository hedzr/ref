module github.com/hedzr/ref

go 1.13

// replace github.com/hedzr/errors v1.1.18

// replace gopkg.in/hedzr/errors.v2 => ../errors

// replace github.com/hedzr/cmdr => ../cmdr

//replace github.com/hedzr/log => ../log

//replace github.com/hedzr/logex => ../logex

// replace github.com/hedzr/cmdr-addons => ../cmdr-addons

// replace github.com/kardianos/service => ../../kardianos/service

// replace github.com/hedzr/go-ringbuf => ../go-ringbuf

//replace github.com/hedzr/assert => ../assert

require (
	github.com/hedzr/assert v0.1.3
	github.com/hedzr/log v0.3.3
	github.com/hedzr/logex v1.3.3
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	gopkg.in/hedzr/errors.v2 v2.1.1
)
