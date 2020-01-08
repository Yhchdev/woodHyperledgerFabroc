package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"net/http"
	"time"
)

func main() {
	router := gin.Default()

	// 定义路由：restful架构风格
	{
		router.POST("/users", userRegister)
		router.GET("/users/:id", queryUser)
		router.DELETE("/users/:id", deleteUser)
		router.GET("/wood/query", queryWood)
		router.POST("/wood/enroll", woodEnroll)
		router.POST("/wood/exchange", woodExchange)
		router.GET("/wood/exchange/history", woodExchangeHistory)
	}

	router.Run() // listen and serve on 0.0.0.0:8080
}

type UserRegisterRequest struct {
	Id   string `form:"id" binding:"required"`
	Name string `form:"name" binding:"required"`
}

// 用户注册
func userRegister(ctx *gin.Context) {
	// 参数处理：从表单获取
	req := new(UserRegisterRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("userRegister", [][]byte{
		[]byte(req.Name),
		[]byte(req.Id),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// 查询用户
func queryUser(ctx *gin.Context) {
	// 参数处理：从url路径上获取
	userId := ctx.Param("id")

	resp, err := channelQuery("queryUser", [][]byte{
		[]byte(userId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

// 用户销户
func deleteUser(ctx *gin.Context) {
	userId := ctx.Param("id")

	resp, err := channelExecute("userDestroy", [][]byte{
		[]byte(userId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// 资产查询
func queryWood(ctx *gin.Context) {

	woodId := ctx.Query("id")
	fcode := ctx.Query("fcode")
	//assetId := ctx.Param("id")

	resp, err := channelQuery("queryWood", [][]byte{
		[]byte(woodId),
		[]byte(fcode),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type WoodEnrollRequest struct {
	WoodId   string `form:"woodid" binding:"required"`
	woodFeatureCode string `form:"woodFeatureCode" binding:"required"`
	Metadata  string `form:"metadata" binding:"required"`
	OwnerId   string `form:"ownerid" binding:"required"`
}

// 资产登记
func woodEnroll(ctx *gin.Context) {
	req := new(WoodEnrollRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	resp, err := channelExecute("woodEnroll", [][]byte{
		[]byte(req.WoodId),
		[]byte(req.woodFeatureCode),
		[]byte(req.Metadata),
		[]byte(req.OwnerId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

type WoodExchangeRequest struct {
	WoodId        string `form:"woodid" binding:"required"`
	OriginOwnerId  string `form:"originownerid" binding:"required"`
	CurrentOwnerId string `form:"currentownerid" binding:"required"`
}

// 资产转让
func woodExchange(ctx *gin.Context) {
	req := new(WoodExchangeRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	resp, err := channelExecute("woodExchange", [][]byte{
		[]byte(req.OriginOwnerId),
		[]byte(req.WoodId),
		[]byte(req.CurrentOwnerId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// 木材历史变更记录
func woodExchangeHistory(ctx *gin.Context) {
	woodId := ctx.Query("woodid")
	queryType := ctx.Query("querytype")

	resp, err := channelQuery("queryWoodHistory", [][]byte{
		[]byte(woodId),
		[]byte(queryType),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

//定义全局变量
var (
	sdk           *fabsdk.FabricSDK
	channelName   = "woodchannel"
	chaincodeName = "woods"
	org           = "org1"
	user          = "Admin"
	//configPath    = "$GOPATH/src/github.com/hyperledger/fabric/wood/application/config.yaml"
	configPath = "./config.yaml"
)

//用配置文件 初始化sdk
func init() {
	var err error
	sdk, err = fabsdk.New(config.FromFile(configPath))
	if err != nil {
		panic(err)
	}
}

// 区块链管理
func manageBlockchain() {
	// 表明身份
	ctx := sdk.Context(fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := resmgmt.New(ctx)
	if err != nil {
		panic(err)
	}

	// 具体操作
	cli.SaveChannel(resmgmt.SaveChannelRequest{}, resmgmt.WithOrdererEndpoint("orderer.wood.com"), resmgmt.WithTargetEndpoints())
}

// 区块链数据查询 账本的查询
func queryBlockchain() {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := ledger.New(ctx)
	if err != nil {
		panic(err)
	}

	resp, err := cli.QueryInfo(ledger.WithTargetEndpoints("peer0.org1.wood.com"))
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	// 读取创世以来所有区块内容 =>区块链浏览器
	// M1:
	cli.QueryBlockByHash(resp.BCI.CurrentBlockHash)

	// M2:
	for i := uint64(0); i <= resp.BCI.Height; i++ {
		cli.QueryBlock(i)
	}
}

// 区块链交互
// 参数1： 函数名
// 参数2： 函数所需参数
func channelExecute(fcn string, args [][]byte) (channel.Response, error) {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	// 状态更新，insert/update/delete
	resp, err := cli.Execute(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.wood.com"))
	if err != nil {
		return channel.Response{}, err
	}

	// 链码事件监听
	go func() {
		// channel
		reg, ccevt, err := cli.RegisterChaincodeEvent(chaincodeName, "eventname")
		if err != nil {
			return
		}
		defer cli.UnregisterChaincodeEvent(reg)

		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for {
			select {
			case evt := <-ccevt:
				fmt.Printf("received event of tx %s: %+v", resp.TransactionID, evt)
			case <-timeoutctx.Done():
				fmt.Println("event timeout, exit!")
				return
			}
		}

		// event
		//eventcli, err := event.New(ctx)
		//if err != nil {
		//	return
		//}

		//eventcli.RegisterChaincodeEvent(chaincodeName, "eventname")
	}()

	// 交易状态事件监听
	go func() {
		eventcli, err := event.New(ctx)
		if err != nil {
			return
		}

		reg, status, err := eventcli.RegisterTxStatusEvent(string(resp.TransactionID))
		defer eventcli.Unregister(reg) // 注册必有注销,注册与注销成对出现

		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for {
			select {
			case evt := <-status:
				fmt.Printf("received event of tx %s: %+v", resp.TransactionID, evt)
			case <-timeoutctx.Done():
				fmt.Println("event timeout, exit!")
				return
			}
		}
	}()

	return resp, nil
}

func channelQuery(fcn string, args [][]byte) (channel.Response, error) {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	// 状态的查询，select
	return cli.Query(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.wood.com"))
}


// 事件监听
func eventHandle() {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := event.New(ctx)
	if err != nil {
		panic(err)
	}

	// 交易状态事件
	// 链码事件 业务事件
	// 区块事件
	reg, blkevent, err := cli.RegisterBlockEvent()
	if err != nil {
		panic(err)
	}
	defer cli.Unregister(reg)

	timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for {
		select {
		case evt := <-blkevent:
			fmt.Printf("received a block", evt)
		case <-timeoutctx.Done():
			fmt.Println("event timeout, exit!")
			return
		}
	}
}




