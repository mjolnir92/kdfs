package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack"
	"github.com/mjolnir92/kdfs/restmsg"
	"github.com/mjolnir92/kdfs/kademlia"
	"github.com/mjolnir92/kdfs/kademliaid"
	"github.com/mjolnir92/kdfs/contact"
	"fmt"
	"net/http"
	"os"
	"log"
	"net"
	"strconv"
)

//var port_rest uint16
var portDHT uint16 = 1200
var joinAddress string
//var dhtAddress string

func init() {
	RootCmd.Flags().StringVarP(&joinAddress, "join", "j", "", "join the a network with a node at address")
	//RootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "the port that the REST API will use")
	//RootCmd.Flags().StringVarP(&dhtAddress, "dht-address", "a", "localhost:9999", "the internet socket that the DHT will use")
}

func main() {
  if err := RootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}

var RootCmd = &cobra.Command{
	Use: "kademlia",
	Short: "starts a kademlia node for kdfs",
	Long: `Starts a kademlia node with a REST API for kdfs.`,
	Run: startServer,
}

var kd *kademlia.T

func startServer(cmd *cobra.Command, args []string) {
	address := getOutboundIP().String() + ":" + strconv.Itoa(int(portDHT))
	kid := kademliaid.NewHash([]byte(address))
	contactMe := contact.New(kid, address)
	kd = kademlia.New(&contactMe)
	go kd.Listen(address)
	if joinAddress != "" {
		kd.Join(joinAddress)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	// prefix everything with /v1
	v1 := router.Group("/v1")
	{
		v1.POST("/store", storeEndpoint)
		v1.GET("/store/:id", getEndpoint)
		v1.POST("/pin/:id", pinEndpoint)
		v1.POST("/unpin/:id", unpinEndpoint)
	}
	router.Run()
}

// NOTE! this IP will be useless if there is NAT between the nodes
// from https://stackoverflow.com/a/37382208
// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    //return conn.LocalAddr().String()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP
}

// POST /store
func storeEndpoint(c *gin.Context) {
	var req restmsg.StoreRequest
	err := c.MustBindWith(&req, binding.MsgPack)
	if err != nil {
		b, err := msgpack.Marshal(restmsg.GenericResponse{Status: http.StatusBadRequest, Message: "Can't read the data"})
		if err != nil {
			panic(fmt.Sprintf("Failed to marshal response: %v", err))
		}
		c.Data(http.StatusOK, binding.MIMEMSGPACK2, b)
		return
	}
	id := kd.KademliaStore(req.File)
	b, err := msgpack.Marshal(restmsg.StoreResponse{Status: http.StatusOK, Message: "Success", ID: id.String()})
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal response: %v", err))
	}
	c.Data(http.StatusOK, binding.MIMEMSGPACK2, b)
}

// GET /store/:id
func getEndpoint(c *gin.Context) {
	var id string = c.Param("id")
	kid := kademliaid.New(id)
	file := kd.Cat(*kid)
	b, err := msgpack.Marshal(restmsg.CatResponse{Status: http.StatusOK, Message: "Success", File: file})
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal response: %v", err))
	}
	c.Data(http.StatusOK, binding.MIMEMSGPACK2, b)
}

// POST /pin/:id
func pinEndpoint(c *gin.Context) {
	var id string = c.Param("id")
	kid := kademliaid.New(id)
	kd.Pin(*kid)
	b, err := msgpack.Marshal(restmsg.GenericResponse{Status: http.StatusOK, Message: "Success"})
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal response: %v", err))
	}
	c.Data(http.StatusOK, binding.MIMEMSGPACK2, b)
}

// DELETE /pin/:id
func unpinEndpoint(c *gin.Context) {
	var id string = c.Param("id")
	kid := kademliaid.New(id)
	kd.Unpin(*kid)
	b, err := msgpack.Marshal(restmsg.GenericResponse{Status: http.StatusOK, Message: "Success"})
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal response: %v", err))
	}
	c.Data(http.StatusOK, binding.MIMEMSGPACK2, b)
}
