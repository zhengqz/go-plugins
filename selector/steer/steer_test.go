package steer_test

import (
	"testing"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-plugins/selector/steer"
)

func TestSteer(t *testing.T) {
	type args struct {
		entropy []string
		nodes   []*registry.Node
		count   int
	}

	type test struct {
		name string
		args args
		want *registry.Node
	}

	node1 := &registry.Node{
		Id: "1",
	}

	node2 := &registry.Node{
		Id: "2",
	}

	node3 := &registry.Node{
		Id: "3",
	}

	nodes := func(n ...*registry.Node) []*registry.Node {
		return n
	}

	tests := []test{
		{
			name: "test single",
			args: args{
				entropy: []string{"a"},
				nodes:   nodes(node1),
				count:   1,
			},
			want: node1,
		},
		{
			name: "test two nodes",
			args: args{
				entropy: []string{"c"},
				nodes:   nodes(node1, node2),
				count:   1,
			},
			want: node2,
		},
		{
			name: "test three nodes",
			args: args{
				entropy: []string{"b"},
				nodes:   nodes(node1, node2, node3),
				count:   1,
			},
			want: node3,
		},
		{
			name: "test three nodes two params",
			args: args{
				entropy: []string{"a", "a"},
				nodes:   nodes(node1, node2, node3),
				count:   1,
			},
			want: node1,
		},
		{
			name: "test three nodes two params two cycles",
			args: args{
				entropy: []string{"a", "a"},
				nodes:   nodes(node1, node2, node3),
				count:   2,
			},
			want: node2,
		},
		{
			name: "test three nodes two params three cycles",
			args: args{
				entropy: []string{"a", "a"},
				nodes:   nodes(node1, node2, node3),
				count:   3,
			},
			want: node3,
		},
		{
			name: "test three nodes two params four cycles",
			args: args{
				entropy: []string{"a", "a"},
				nodes:   nodes(node1, node2, node3),
				count:   4,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := steer.Strategy(tt.args.entropy...)
			next := getNext(fn, tt.args.nodes)

			if next == nil {
				t.Error("got nil next")
				return
			}

			var err error
			var got *registry.Node
			for i := 0; i < tt.args.count; i++ {
				got, err = next()
				if err != nil {
					t.Errorf("got error with next %v", err)
					return
				}
			}

			if got != tt.want {
				t.Errorf("Steer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getNext(fn client.CallOption, nodes []*registry.Node) selector.Next {
	co := &client.CallOptions{}
	fn(co)

	if len(co.SelectOptions) != 1 {
		return nil
	}

	opt := co.SelectOptions[0]

	so := &selector.SelectOptions{}
	opt(so)

	if so.Strategy == nil {
		return nil
	}

	return so.Strategy([]*registry.Service{
		{
			Nodes: nodes,
		},
	})
}
