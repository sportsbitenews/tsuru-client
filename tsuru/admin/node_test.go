// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package admin

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/ajg/form"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"github.com/tsuru/tsuru/provision"
	"gopkg.in/check.v1"
)

func (s *S) TestAddNodeCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"pool=poolTest", "address=http://localhost:8080"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			var params provision.AddNodeOptions
			dec := form.NewDecoder(nil)
			dec.IgnoreUnknownKeys(true)
			err = dec.DecodeValues(&params, req.Form)
			c.Assert(err, check.IsNil)
			u := strings.HasSuffix(req.URL.Path, "/1.2/node")
			method := req.Method == "POST"
			address := params.Metadata["address"] == "http://localhost:8080"
			pool := params.Metadata["pool"] == "poolTest"
			register := !params.Register
			return u && method && register && address && pool
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := AddNodeCmd{register: false}
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully registered.\n")
}

func (s *S) TestAddNodeWithErrorCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{
		Args:   []string{"pool=poolTest", "address=http://localhost:8080"},
		Stdout: &buf, Stderr: &buf,
	}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{"error": "some err"}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			var params provision.AddNodeOptions
			dec := form.NewDecoder(nil)
			dec.IgnoreUnknownKeys(true)
			err = dec.DecodeValues(&params, req.Form)
			c.Assert(err, check.IsNil)
			u := strings.HasSuffix(req.URL.Path, "/1.2/node")
			method := req.Method == "POST"
			address := params.Metadata["address"] == "http://localhost:8080"
			pool := params.Metadata["pool"] == "poolTest"
			register := !params.Register
			return u && method && register && address && pool
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := AddNodeCmd{register: false}
	err := cmd.Run(&context, client)
	c.Assert(err.Error(), check.Equals, "some err")
}

func (s *S) TestRemoveNodeFromTheSchedulerCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:8080"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			u := strings.HasSuffix(req.URL.Path, "/1.2/node/http://localhost:8080")
			raw := req.URL.RawQuery == "no-rebalance=false"
			method := req.Method == "DELETE"
			return u && method && raw
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := RemoveNodeCmd{}
	cmd.Flags().Parse(true, []string{"-y"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully removed.\n")
}

func (s *S) TestRemoveNodeFromTheSchedulerWithDestroyCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:8080"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			u := strings.HasSuffix(req.URL.Path, "/1.2/node/http://localhost:8080")
			raw := req.URL.RawQuery == "no-rebalance=false&remove-iaas=true"
			method := req.Method == "DELETE"
			return u && method && raw
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := RemoveNodeCmd{}
	cmd.Flags().Parse(true, []string{"-y", "--destroy"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully removed.\n")
}

func (s *S) TestRemoveNodeFromTheSchedulerWithDestroyCmdRunConfirmation(c *check.C) {
	var stdout bytes.Buffer
	context := cmd.Context{
		Args:   []string{"http://localhost:8080"},
		Stdout: &stdout,
		Stdin:  strings.NewReader("n\n"),
	}
	command := RemoveNodeCmd{}
	err := command.Run(&context, nil)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Are you sure you sure you want to remove \"http://localhost:8080\" from cluster? (y/n) Abort.\n")
}

func (s *S) TestRemoveNodeFromTheSchedulerWithNoRebalanceCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:8080"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			u := strings.HasSuffix(req.URL.Path, "/1.2/node/http://localhost:8080")
			raw := req.URL.RawQuery == "no-rebalance=true"
			method := req.Method == "DELETE"
			return u && method && raw
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := RemoveNodeCmd{}
	cmd.Flags().Parse(true, []string{"-y", "--no-rebalance"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully removed.\n")
}

func (s *S) TestListNodesCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{
	"machines": [{"Id": "m-id-1", "Address": "localhost2"}],
	"nodes": [
		{"Address": "http://localhost1:8080", "Status": "disabled", "Metadata": {"meta1": "foo", "meta2": "bar"}},
		{"Address": "http://localhost2:9090", "Status": "ready"}
	]
}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := (&ListNodesCmd{}).Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `+------------------------+---------+----------+-----------+
| Address                | IaaS ID | Status   | Metadata  |
+------------------------+---------+----------+-----------+
| http://localhost1:8080 |         | disabled | meta1=foo |
|                        |         |          | meta2=bar |
+------------------------+---------+----------+-----------+
| http://localhost2:9090 | m-id-1  | ready    |           |
+------------------------+---------+----------+-----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestListNodesCmdRunWithFilters(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{
	"machines": [{"Id": "m-id-1", "Address": "localhost2"}],
	"nodes": [
		{"Address": "http://localhost1:8080", "Status": "disabled", "Metadata": {"meta1": "foo", "meta2": "bar"}},
		{"Address": "http://localhost2:8089", "Status": "disabled"},
		{"Address": "http://localhost2:9090", "Status": "disabled", "Metadata": {"meta1": "foo"}}
	]
}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := ListNodesCmd{}
	cmd.Flags().Parse(true, []string{"--filter", "meta1=foo", "-f", "meta2=bar"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `+------------------------+---------+----------+-----------+
| Address                | IaaS ID | Status   | Metadata  |
+------------------------+---------+----------+-----------+
| http://localhost1:8080 |         | disabled | meta1=foo |
|                        |         |          | meta2=bar |
+------------------------+---------+----------+-----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestListNodesCmdRunEmptyAll(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := (&ListNodesCmd{}).Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `+---------+---------+--------+----------+
| Address | IaaS ID | Status | Metadata |
+---------+---------+--------+----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestListNodesCmdRunNoContent(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{}`, Status: http.StatusNoContent},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := (&ListNodesCmd{}).Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `+---------+---------+--------+----------+
| Address | IaaS ID | Status | Metadata |
+---------+---------+--------+----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestListNodesCmdRunWithFlagQ(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{
	"machines": [{"Id": "m-id-1", "Address": "localhost2"}],
	"nodes": [
		{"Address": "http://localhost1:8080", "Status": "disabled", "Metadata": {"meta1": "foo", "meta2": "bar"}},
		{"Address": "http://localhost1:8989", "Status": "disabled", "Metadata": {"meta2": "bar"}},
		{"Address": "http://localhost2:9090", "Status": "ready"}
	]
}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := ListNodesCmd{}
	cmd.Flags().Parse(true, []string{"-q", "-f", "meta2=bar"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := "http://localhost1:8080\nhttp://localhost1:8989\n"
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestUpdateNodeCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:1111", "x=y", "y=z"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			u := strings.HasSuffix(req.URL.Path, "/1.2/node")
			method := req.Method == "PUT"
			var params provision.UpdateNodeOptions
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			dec := form.NewDecoder(nil)
			dec.IgnoreUnknownKeys(true)
			err = dec.DecodeValues(&params, req.Form)
			c.Assert(err, check.IsNil)
			address := params.Address == "http://localhost:1111"
			x := params.Metadata["x"] == "y"
			y := params.Metadata["y"] == "z"
			disabled := !params.Disable
			return u && method && address && x && y && disabled
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := UpdateNodeCmd{}
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully updated.\n")
}

func (s *S) TestUpdateNodeToDisableCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:1111", "x=y", "y=z"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			var params provision.UpdateNodeOptions
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			dec := form.NewDecoder(nil)
			dec.IgnoreUnknownKeys(true)
			err = dec.DecodeValues(&params, req.Form)
			c.Assert(err, check.IsNil)
			return params.Disable
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cm := UpdateNodeCmd{}
	cm.Flags().Parse(true, []string{"--disable"})
	err := cm.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully updated.\n")
}

func (s *S) TestUpdateNodeToEnabledCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:1111", "x=y", "y=z"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			var params provision.UpdateNodeOptions
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			dec := form.NewDecoder(nil)
			dec.IgnoreUnknownKeys(true)
			err = dec.DecodeValues(&params, req.Form)
			c.Assert(err, check.IsNil)
			return params.Enable
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cm := UpdateNodeCmd{}
	cm.Flags().Parse(true, []string{"--enable"})
	err := cm.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node successfully updated.\n")
}

func (s *S) TestUpdateNodeToEnabledDisableCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"http://localhost:1111", "x=y", "y=z"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{
			Message: "You can't make a node enable and disable at the same time.",
			Status:  http.StatusBadRequest,
		},
		CondFunc: func(req *http.Request) bool {
			enabled := req.FormValue("enable") == "true"
			disabled := req.FormValue("disable") == "true"
			return enabled && disabled
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cm := UpdateNodeCmd{}
	cm.Flags().Parse(true, []string{"--enable", "--disable"})
	err := cm.Run(&context, client)
	c.Assert(err, check.NotNil)
}

func (s *S) TestGetNodeHealingConfigCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{
"": {"enabled": true, "maxunresponsivetime": 2},
"p1": {"enabled": false, "maxunresponsivetime": 2, "maxunresponsivetimeinherited": true},
"p2": {"enabled": true, "maxunresponsivetime": 3, "enabledinherited": true}
}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/healing/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	healing := &GetNodeHealingConfigCmd{}
	err := healing.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `Default:
+------------------------+----------+
| Config                 | Value    |
+------------------------+----------+
| Enabled                | true     |
| Max unresponsive time  | 2s       |
| Max time since success | disabled |
+------------------------+----------+

Pool "p1":
+------------------------+----------+-----------+
| Config                 | Value    | Inherited |
+------------------------+----------+-----------+
| Enabled                | false    | false     |
| Max unresponsive time  | 2s       | true      |
| Max time since success | disabled | false     |
+------------------------+----------+-----------+

Pool "p2":
+------------------------+----------+-----------+
| Config                 | Value    | Inherited |
+------------------------+----------+-----------+
| Enabled                | true     | true      |
| Max unresponsive time  | 3s       | false     |
| Max time since success | disabled | false     |
+------------------------+----------+-----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestGetNodeHealingConfigCmdEmpty(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.2/healing/node"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	healing := &GetNodeHealingConfigCmd{}
	err := healing.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `Default:
+------------------------+----------+
| Config                 | Value    |
+------------------------+----------+
| Enabled                | false    |
| Max unresponsive time  | disabled |
| Max time since success | disabled |
+------------------------+----------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestDeleteNodeHealingConfigCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			req.ParseForm()
			return req.URL.Path == "/1.2/healing/node" && req.Method == "DELETE" &&
				req.Form.Get("name") == "Enabled" && req.Form.Get("pool") == "p1"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	healing := &DeleteNodeHealingConfigCmd{}
	healing.Flags().Parse(true, []string{"--enabled", "--pool", "p1", "-y"})
	err := healing.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node healing configuration successfully removed.\n")
}

func (s *S) TestSetNodeHealingConfigCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: `{}`, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			req.ParseForm()
			c.Assert(req.Form, check.DeepEquals, url.Values{
				"pool":                []string{"p1"},
				"MaxUnresponsiveTime": []string{"10"},
				"Enabled":             []string{"false"},
			})
			return req.URL.Path == "/1.2/healing/node" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	healing := &SetNodeHealingConfigCmd{}
	healing.Flags().Parse(true, []string{"--pool", "p1", "--disable", "--max-unresponsive", "10"})
	err := healing.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node healing configuration successfully updated.\n")
}
