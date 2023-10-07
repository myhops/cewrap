package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type options struct {
	downstream    string
	port          string
	sink          string
	source        string
	dataschema    string
	changeMethods []string
	typePrefix    string
}

func (o *options) setChangeMethods(mlist string) {
	o.changeMethods = nil
	o.appendChangeMethods(mlist)
}

func (o *options) appendChangeMethods(mlist string) {
	m := strings.Split(mlist, ",")
	o.changeMethods = append(o.changeMethods, m...)
}

// args is without the program name as first parameter
func getOptionsFrom(args, env []string) (*options, error) {
	opts := &options{}
	opts.getEnv(env)
	if err := opts.parseArgs(args); err != nil {
		return nil, err
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return opts, nil
}

func getOptions() (*options, error) {
	return getOptionsFrom(os.Args[1:], os.Environ())
}

func envVar(s string) (key string, value string) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", ""
	}
	key = s[:i]
	value = s[i+1:]
	return key, value
}

func (o *options) getEnv(env []string) error {
	for _, ev := range env {
		k, v := envVar(ev)
		if k == "" {
			continue
		}
		switch k {
		case "K_SINK", "CEW_SINK":
			o.sink = v
		case "PORT":
			o.port = v
		case "CEW_DOWNSTREAM":
			o.downstream = v
		case "CEW_SOURCE":
			o.source = v
		case "CEW_TYPE_PREFIX":
			o.typePrefix = v
		case "CEW_DATASCHEMA":
			o.dataschema = v
		case "CEW_CHANGE_METHODS":
			o.setChangeMethods(v)
		case "CEW_EXTRA_METHODS":
			o.appendChangeMethods(v)
		}
	}
	return nil
}

func (o *options) parseArgs(args []string) error {
	fs := flag.NewFlagSet("root", flag.ExitOnError)

	downstream := fs.String("downstream", "", "downstream service")
	port := fs.String("port", "", "port to listen on")
	sink := fs.String("sink", "", "url of the event sink")
	typePrefix := fs.String("type", "", "type prefix")
	dataschema := fs.String("dataschema", "", "dataschema")
	changeMethods := fs.String("change-methods", "", "override the default change methods, do not use together with extra-methods")
	extraMethods := fs.String("extra-methods", "", "additional methods to trigger an event on, do not use together with change-methods")

	if err := fs.Parse(args); err != nil {
		return err
	}
	// copy the set vars to options.
	if *downstream != "" {
		o.downstream = *downstream
	}
	if *port != "" {
		o.port = *port
	}
	if *sink != "" {
		o.sink = *sink
	}
	if *typePrefix != "" {
		o.typePrefix = *typePrefix
	}
	if *dataschema != "" {
		o.dataschema = *dataschema
	}
	if *changeMethods != "" {
		o.setChangeMethods(*changeMethods)
	}
	if *extraMethods != "" {
		o.appendChangeMethods(*extraMethods)
	}

	return nil
}

func (o *options) validate() error {
	var errs []error

	// Check the urls.
	if o.downstream != "" {
		if _, err := url.Parse(o.downstream); err != nil {
			errs = append(errs, fmt.Errorf("error parsing downstream: %w", err))
		}
	} else {
		errs = append(errs, errors.New("downstream not set"))
	}

	if o.sink != "" {
		if _, err := url.Parse(o.sink); err != nil {
			errs = append(errs, fmt.Errorf("error parsing sink: %w", err))
		}
	} else {
		errs = append(errs, errors.New("sink not set"))
	}

	// Check if port is set and numeric
	if o.port != "" {
		if _, err := strconv.Atoi(o.port); err != nil {
			errs = append(errs, fmt.Errorf("port is not numeric: %w", err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	// Set port to default.
	if o.port == "" {
		o.port = "8080"
	}
	return nil
}
