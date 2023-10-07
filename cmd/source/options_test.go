package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	cases := []struct {
		name string
		env  []string
		want *options
	}{
		{
			name: "env ok",
			env: []string{
				"K_SINK=http://example.com/sink",
				"PORT=9090",
				"CEW_DOWNSTREAM=http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
			},
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			opts := &options{}
			err := opts.getEnv(cc.env)
			if err != nil {
				t.Fatalf("getEnv error: %s", err)
			}
			if cc.want.downstream != opts.downstream {
				t.Errorf("downstream, want %s, got %s", cc.want.downstream, opts.downstream)
			}
			if cc.want.sink != opts.sink {
				t.Errorf("sink, want %s, got %s", cc.want.sink, opts.sink)
			}
			if cc.want.port != opts.port {
				t.Errorf("port, want %s, got %s", cc.want.port, opts.port)
			}
		})

	}
}

func TestArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want *options
	}{
		{
			name: "args ok 1",
			args: []string{
				"-sink=http://example.com/sink",
				"-port=9090",
				"-downstream=http://example.com/downstream",
				"-change-methods=POST,PUT",
			},
			want: &options{
				sink:          "http://example.com/sink",
				port:          "9090",
				downstream:    "http://example.com/downstream",
				changeMethods: []string{"POST", "PUT"},
			},
		},
		{
			name: "args ok 2",
			args: []string{
				"-sink", "http://example.com/sink",
				"-port", "9090",
				"-downstream", "http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
			},
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			opts := &options{}
			err := opts.parseArgs(cc.args)
			if err != nil {
				t.Fatalf("getEnv error: %s", err)
			}
			if cc.want.downstream != opts.downstream {
				t.Errorf("downstream, want %s, got %s", cc.want.downstream, opts.downstream)
			}
			if cc.want.sink != opts.sink {
				t.Errorf("sink, want %s, got %s", cc.want.sink, opts.sink)
			}
			if cc.want.port != opts.port {
				t.Errorf("port, want %s, got %s", cc.want.port, opts.port)
			}
			assert.Equal(t, cc.want.changeMethods, opts.changeMethods)
		})
	}
}

func TestArgsEnv(t *testing.T) {
	cases := []struct {
		name string
		args []string
		env  []string
		want *options
	}{
		{
			name: "args",
			env:  []string{},
			args: []string{
				"-sink=http://example.com/sink",
				"-port=9090",
				"-downstream=http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
			},
		},
		{
			name: "one",
			env: []string{
				"K_SINK=http://example.com/sink",
				"PORT=9090",
				"CEW_DOWNSTREAM=http://example.com/downstream",
			},
			args: []string{
				"-sink=http://example.com/sink",
				"-port=9090",
				"-downstream=http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
			},
		},
		{
			name: "two",
			env: []string{
				"K_SINK=http://example.com/sinkenv",
				"PORT=7070",
				"CEW_DOWNSTREAM=http://example.com/downstreamenv",
			},
			args: []string{
				"-sink", "http://example.com/sink",
				"-port", "9090",
				"-downstream", "http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
			},
		},
		{
			name: "more",
			env: []string{
				"K_SINK=http://example.com/sinkenv",
				"PORT=7070",
				"CEW_DOWNSTREAM=http://example.com/downstreamenv",
				"CEW_CHANGE_METHODS=PUT,POST",
				"CEW_DATASCHEMA=dataschema",
				"CEW_SOURCE=source",
				"CEW_TYPE_PREFIX=typeprefix",
			},
			args: []string{
				"-sink", "http://example.com/sink",
				"-port", "9090",
				"-downstream", "http://example.com/downstream",
			},
			want: &options{
				sink:       "http://example.com/sink",
				port:       "9090",
				downstream: "http://example.com/downstream",
				changeMethods: []string{"PUT", "POST"},
				source: "source",
				dataschema: "dataschema",
				typePrefix: "typeprefix",
			},
		},

	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			opts, err := getOptionsFrom(cc.args, cc.env)
			assert.NoError(t, err)
			assert.Equal(t, cc.want, opts)
		})
	}
}
