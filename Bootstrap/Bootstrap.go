package Bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/amidmm/MyChain/Transaction"

	"github.com/amidmm/MyChain/Messages"

	"github.com/amidmm/MyChain/Account"
	"github.com/amidmm/MyChain/Blockchain"
	"github.com/amidmm/MyChain/Config"
	"github.com/amidmm/MyChain/Tangle"
)

var ctx = context.Background()
var packetChan chan *msg.Packet
var bc *Blockchain.Blockchain
var t *Tangle.Tangle
var user *Account.User
var advertiser chan *msg.Packet

func Run() (context.Context, chan *msg.Packet, *Blockchain.Blockchain, *Tangle.Tangle, *Account.User, chan *msg.Packet, error) {
	if ctx != nil && bc != nil && t != nil && packetChan != nil && user != nil {
		return ctx, packetChan, bc, t, user, advertiser, nil
	}

	var user *Account.User
	user, err := Account.LoadUser(Config.Username)
	if err != nil {
		Account.CreateUser(Config.Username, nil, nil, false, 0)
		user, _ = Account.LoadUser(Config.Username)
	}

	Transaction.ThisUserDb = user.DB
	Transaction.ThisUserAddr, _ = user.PubKey.Bytes()
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
	advertiser = make(chan *msg.Packet, 100)
	return ctx, packetChan, bc, t, user, advertiser, nil
}
