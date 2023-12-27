package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/myhops/cewrap"
)

type options struct {
	downstream string
	port       string
	sink       string
	source     string
	dataschema string
	typePrefix string
	pathPrefix string
	logFormat  string
	logLevel   string

	changeMethods    []string
	changeMethodsSet bool
}

// setChangeMethods sets the change methods from mlist.
// 	mlist contains the methods separated by a comma.
func (o *options) setChangeMethods(mlist string) {
	if len(o.changeMethods) == 0 {
		o.appendChangeMethods(mlist)
		o.changeMethodsSet = true
	}
}

// appendChangeMethods appends the methods in the list to the change methods.
func (o *options) appendChangeMethods(mlist string) {
	m := strings.Split(mlist, ",")
	o.changeMethods = append(o.changeMethods, m...)
}

// getOptionsFrom gets the options from the cli arguments and the environment vars.
func getOptionsFrom(args, env []string) (*options, error) {
	opts := &options{}
	// Env fist.
	opts.getEnv(env)
	// Args overrule
	if err := opts.parseArgs(args); err != nil {
		return nil, err
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return opts, nil
}

// getOptions gets the options from env and cli args.
func getOptions() (*options, error) {
	return getOptionsFrom(os.Args[1:], os.Environ())
}

// envVar takes a key=value string and returns the key and the value.
// The value can contain the = char.
func envVar(s string) (key string, value string) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", ""
	}
	key = s[:i]
	value = s[i+1:]
	return key, value
}

// getEnv parses the env vars into the options.
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
		case "CEW_PATH_PREFIX":
			o.pathPrefix = v
		case "CEW_CHANGE_METHODS":
			o.setChangeMethods(v)
		case "CEW_EXTRA_METHODS":
			o.appendChangeMethods(v)
		case "CEW_LOG_FORMAT":
			o.logFormat = v
		case "CEW_LOG_LEVEL":
			o.logLevel = v
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
	pathPrefix := fs.String("path-prefix", "", "path prefix is removed from the subject")
	changeMethods := fs.String("change-methods", "", "override the default change methods, do not use together with extra-methods")
	extraMethods := fs.String("extra-methods", "", "additional methods to trigger an event on, do not use together with change-methods")
	logFormat := fs.String("log-format", "", "log format, json or text")
	logLevel := fs.String("log-level", "", "log level, debug, info, warn, error")

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
	if *pathPrefix != "" {
		o.pathPrefix = *pathPrefix
	}
	if *changeMethods != "" {
		o.setChangeMethods(*changeMethods)
	}
	if *extraMethods != "" {
		o.appendChangeMethods(*extraMethods)
	}
	if *logFormat != "" {
		o.logFormat = *logFormat
	}
	if *logLevel != "" {
		o.logLevel = *logLevel
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

	// Append the default change methods.
	if !o.changeMethodsSet {
		o.changeMethods = append(o.changeMethods, cewrap.DefaultChangeMethods...)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	// Set port to default.
	if o.port == "" {
		o.port = "8080"
	}
	// Check logformat
	switch o.logFormat {
	case "text", "json":
		break
	default:
		o.logFormat = "text"
	}
	// Check logformat
	switch o.logLevel {
	case "debug", "info", "warn", "error":
		break
	default:
		o.logLevel = "info"
	}

	return nil
}

func (o *options) getSourceOptions() ([]cewrap.SourceOption, error) {
	var so []cewrap.SourceOption

	// create the sink
	sink, err := client.NewHTTP(cloudevents.WithTarget(o.sink))
	if err != nil {
		return nil, err
	}

	so = append(so,
		cewrap.WithDownstream(o.downstream),
		cewrap.WithChangeMethods(o.changeMethods),
		cewrap.WithSource(o.source),
		cewrap.WithDataschema(o.dataschema),
		cewrap.WithTypePrefix(o.typePrefix),
		cewrap.WithPathPrefix(o.pathPrefix),
		cewrap.WithSink(sink),
	)
	return so, nil
}
