package Statistics

import (
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

type State struct {
	SystemBootTime  time.Time
	SystemShutDown  time.Time
	TotalPacket     int
	LastTotalPacket time.Time
	OutPacket       int
	LastOutPacket   time.Time
	TotalBlock      int
	LastTotalBlock  time.Time
	OutBlock        int
	LastOutBlock    time.Time
	TotalTangle     int
	LastTotalTangle time.Time
	OutTangle       int
	LastOutTangle   time.Time
}

var SystemState State

func init() {
	SystemState.SystemBootTime = time.Now()
}

func (s *State) NewTotalBlock() {
	s.TotalBlock++
	s.TotalPacket++
	s.LastTotalBlock = time.Now()
	s.LastTotalPacket = time.Now()
}
func (s *State) NewTotalTangle() {
	s.TotalTangle++
	s.TotalPacket++
	s.LastTotalTangle = time.Now()
	s.LastTotalPacket = time.Now()
}
func (s *State) NewOutBlock() {
	s.OutBlock++
	s.OutPacket++
	s.LastOutBlock = time.Now()
	s.LastOutPacket = time.Now()
}
func (s *State) NewOutTangle() {
	s.OutTangle++
	s.OutPacket++
	s.LastOutTangle = time.Now()
	s.LastOutPacket = time.Now()
}

func (s *State) Pretty() {
	data := [][]string{
		[]string{"In Block", strconv.Itoa(s.TotalBlock - s.OutBlock), s.LastTotalBlock.String()},
		[]string{"Out Block", strconv.Itoa(s.OutBlock), s.LastOutBlock.String()},
		[]string{"Total Block", strconv.Itoa(s.TotalBlock), ""},
		[]string{"In Tangle", strconv.Itoa(s.TotalTangle - s.OutTangle), s.LastTotalTangle.String()},
		[]string{"Out Tangle", strconv.Itoa(s.OutTangle), s.LastOutTangle.String()},
		[]string{"Total Tangle", strconv.Itoa(s.TotalTangle), ""},
		[]string{"In Packet", strconv.Itoa(s.TotalPacket - s.OutPacket), s.LastTotalPacket.String()},
		[]string{"Out Packet", strconv.Itoa(s.OutPacket), s.LastOutPacket.String()},
		[]string{"Total Packet", strconv.Itoa(s.TotalPacket), ""},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Number Of", "Timestamp"})
	table.AppendBulk(data)
	table.Append([]string{"", "System Boot", s.SystemBootTime.String()})
	table.Append([]string{"", "Uptime", time.Since(s.SystemBootTime).String()})
	table.SetRowLine(true)
	table.Render()
}
