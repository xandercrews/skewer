// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/stephane-martin/skewer/javascript"
	"github.com/stephane-martin/skewer/model"

	"github.com/spf13/cobra"
)

var testjsCmd = &cobra.Command{
	Use:   "testjs",
	Short: "Debugging stuff for the Ecmascript VM",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("testjs called")
		logger := log15.New()
		ffunc := `function FilterMessages(m) { m.Message="bla"; return FILTER.PASS; }`
		tfunc := `function Topic(m) { return "topic-" + m.Appname; }`
		pfunc := `function PartitionNumber(m) {return 4; }`
		env := javascript.NewFilterEnvironment(ffunc, tfunc, "", "", "", pfunc, logger)
		m := &model.SyslogMessage{}
		m.TimeReportedNum = time.Now().UnixNano()
		m.TimeGeneratedNum = time.Now().Add(time.Hour).UnixNano()
		m.Facility = 5
		m.Severity = 2
		m.Priority = 11
		m.Version = 3
		m.HostName = "myhostname"
		m.ProcId = "myprocid"
		m.MsgId = "mymsgid"
		m.AppName = "myapp"
		m.Message = "orig message"
		m.SetProperty("foo", "zog", "zogzog")
		m.SetProperty("bar", "zobi", "la mouche")
		result, err := env.FilterMessage(m)
		fmt.Println(err)
		fmt.Println(result)

		topic, errs := env.Topic(m)
		fmt.Println(errs)
		fmt.Println(topic)

		pnumber, errs := env.PartitionNumber(m)
		fmt.Println(errs)
		fmt.Println(pnumber)

	},
}

func init() {
	RootCmd.AddCommand(testjsCmd)
}
