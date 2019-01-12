package Bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/amidmm/MyChain/Messages"

	"github.com/amidmm/MyChain/Account"
	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Config"
	"github.com/amidmm/MyChain/Tangle"
	"github.com/amidmm/MyChain/Utils"
)

var ctx = context.Background()
var packetChan chan *msg.Packet
var bc *Blockchain.Blockchain
var t *Tangle.Tangle
var user *Account.User

func Run() (context.Context, chan *msg.Packet, *Blockchain.Blockchain, *Tangle.Tangle, *Account.User, error) {
	if ctx != nil && bc != nil && t != nil && packetChan != nil && user != nil {
		return ctx, packetChan, bc, t, user, nil
	}
	//TODO: temp
	Utils.RemoveExistingDB()
	var user *Account.User
	user, err := Account.LoadUser(Config.Username)
	if err != nil {
		Account.CreateUser(Config.Username, nil, nil, false, 0)
		user, _ = Account.LoadUser(Config.Username)
	}
	bc, err := Blockchain.NewBlockchain()
	if err != nil {
		//TODO: do sth
	}
	bc.InitBlockchain()
	t, err = Tangle.NewTangle(bc)
	if err != nil {
		//TODO: do sth
	}
	t.InitTangle()
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
		log.Println("\033[41m Shuting down....\033[41m")
		cancel()
	}()
	packetChan = make(chan *msg.Packet)
	return ctx, packetChan, bc, t, user, nil
}
