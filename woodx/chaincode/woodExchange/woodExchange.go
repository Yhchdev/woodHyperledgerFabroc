package main

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type AssertsExchangeCC struct{}

const (
	originOwner = "originOwnerPlaceholder"
)


/*
	1.定义系统实体对象
 */

// 用户 结构体
type User struct{
	Name string `json:"name"` //messagepack || protobuf
	Id string `json:"id"`   //唯一标识
	Woods []string `json:"woods"`
}

// 木材
type Wood struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	//Metadata map[string]string `json:"metadata"` // 特殊属性
	Metadata string `json:"metadata"` // 特殊属性
}

// 木材交易
type WoodHistory struct {
	WoodId         string `json:"wood_id"`
	OriginOwnerId  string `json:"origin_owner_id"`  // 木材的原始拥有者
	CurrentOwnerId string `json:"current_owner_id"` // 变更后当前的拥有者
}

func constructUserKey(userId string) string {
	return fmt.Sprintf("user_%s", userId)
}

func constructWoodKey(woodId string) string {
	return fmt.Sprintf("wood_%s", woodId)
}

/*
	链码编写套路：
		1.检查参数的个数是否正确
		2.检验参数的正确性
 */

// 用户开户
func userRegister(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 2 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	name := args[0]
	id := args[1]
	if name == "" || id == "" {
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	if userBytes, err := stub.GetState(constructUserKey(id)); err == nil && len(userBytes) != 0 {
		return shim.Error("user already exist")
	}

	// 套路4：写入状态
	user := &User{
		Name:  name,
		Id:    id,
		Woods: make([]string, 0),
	}

	// 序列化对象
	userBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error %s", err))
	}

	if err := stub.PutState(constructUserKey(id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("put user error %s", err))
	}

	// 成功返回
	return shim.Success(nil)
}

// 用户销户
func userDestroy(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 1 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	id := args[0]
	if id == "" {
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	userBytes, err := stub.GetState(constructUserKey(id))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not found")
	}

	// 套路4：写入状态 (删除DelState)
	if err := stub.DelState(constructUserKey(id)); err != nil {
		return shim.Error(fmt.Sprintf("delete user error: %s", err))
	}

	// 删除用户名下的木材资产
	//user := new(User)
	//if err := json.Unmarshal(userBytes, user); err != nil {
	//	return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	//}
	//for _, assetid := range user.Assets {
	//	if err := stub.DelState(constructAssetKey(assetid)); err != nil {
	//		return shim.Error(fmt.Sprintf("delete asset error: %s", err))
	//	}
	//}

	return shim.Success(nil)
}

// 木材登记
func woodEnroll(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 4 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	woodId := args[0]
	woodFeatureCode := args[1]
	metadata := args[2]
	ownerId := args[3]
	if woodId == "" || woodFeatureCode == "" || ownerId == "" {
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	userBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not found")
	}

	if assetBytes, err := stub.GetState(constructWoodKey(woodId)); err == nil && len(assetBytes) != 0 {
		return shim.Error("asset already exist")
	}

	// 套路4：写入状态
	// 1. 写入资产对象 2. 更新用户对象 3. 写入资产变更记录
	wood := &wood{
		Name:     woodName,
		Id:       woodId,
		Metadata: metadata,
	}
	woodBytes, err := json.Marshal(wood)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal wood error: %s", err))
	}
	if err := stub.PutState(constructWoodKey(woodId), woodBytes); err != nil {
		return shim.Error(fmt.Sprintf("save wood error: %s", err))
	}

	user := new(User)
	// 反序列化user
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}
	user.Woods = append(user.Woods, woodId)
	// 序列化user
	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error: %s", err))
	}
	if err := stub.PutState(constructUserKey(user.Id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("update user error: %s", err))
	}

	// 资产变更历史
	history := &WoodHistory{
		WoodId:         woodId,
		OriginOwnerId:  originOwner,
		CurrentOwnerId: ownerId,
	}
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal assert history error: %s", err))
	}

	historyKey, err := stub.CreateCompositeKey("history", []string{
		woodId,
		originOwner,
		ownerId,
	})
	if err != nil {
		return shim.Error(fmt.Sprintf("create key error: %s", err))
	}

	if err := stub.PutState(historyKey, historyBytes); err != nil {
		return shim.Error(fmt.Sprintf("save wood history error: %s", err))
	}

	return shim.Success(nil)
}

// 资产转让
func woodExchange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 3 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	ownerId := args[0]
	woodId := args[1]
	currentOwnerId := args[2]
	if ownerId == "" || woodId == "" || currentOwnerId == "" {
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	// 确定交易双方为合法（存在）的人
	originOwnerBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(originOwnerBytes) == 0 {
		return shim.Error("user not found")
	}

	currentOwnerBytes, err := stub.GetState(constructUserKey(currentOwnerId))
	if err != nil || len(currentOwnerBytes) == 0 {
		return shim.Error("user not found")
	}

	assetBytes, err := stub.GetState(constructWoodKey(woodId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("asset not found")
	}

	// 校验原始拥有者确实拥有当前变更的木材
	originOwner := new(User)
	// 反序列化user
	if err := json.Unmarshal(originOwnerBytes, originOwner); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}
	aidexist := false
	//遍历这个人名下的木材资产
	for _, aid := range originOwner.Woods {
		if aid == woodId {
			aidexist = true
			break
		}
	}
	if !aidexist {
		return shim.Error("asset owner not match")
	}

	// 套路4：写入状态
	// 1. 拥有者删除资产id 2. 新拥有者加入资产id 3. 资产变更记录
	woodIds := make([]string, 0)
	for _, aid := range originOwner.Woods {
		if aid == woodId {
			continue
		}
		// 未转移的木材
		woodIds = append(woodIds, aid)
	}
	originOwner.Woods = woodIds

	originOwnerBytes, err = json.Marshal(originOwner)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error: %s", err))
	}
	if err := stub.PutState(constructUserKey(ownerId), originOwnerBytes); err != nil {
		return shim.Error(fmt.Sprintf("update user error: %s", err))
	}

	// 当前拥有者插入资产id
	currentOwner := new(User)
	// 反序列化user
	if err := json.Unmarshal(currentOwnerBytes, currentOwner); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}
	currentOwner.Woods = append(currentOwner.Woods, woodId)

	currentOwnerBytes, err = json.Marshal(currentOwner)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error: %s", err))
	}
	if err := stub.PutState(constructUserKey(currentOwnerId), currentOwnerBytes); err != nil {
		return shim.Error(fmt.Sprintf("update user error: %s", err))
	}

	// 插入资产变更记录
	history := &WoodHistory{
		WoodId:         woodId,
		OriginOwnerId:  ownerId,
		CurrentOwnerId: currentOwnerId,
	}
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal wood history error: %s", err))
	}

	historyKey, err := stub.CreateCompositeKey("history", []string{
		woodId,
		ownerId,
		currentOwnerId,
	})
	if err != nil {
		return shim.Error(fmt.Sprintf("create key error: %s", err))
	}

	if err := stub.PutState(historyKey, historyBytes); err != nil {
		return shim.Error(fmt.Sprintf("save wood history error: %s", err))
	}

	return shim.Success(nil)
}

// 用户查询
func queryUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 1 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	ownerId := args[0]
	if ownerId == "" {
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	userBytes, err := stub.GetState(constructUserKey(ownerId))
	if err != nil || len(userBytes) == 0 {
		return shim.Error("user not found")
	}

	return shim.Success(userBytes)
}

// 木材查询
func queryWood(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 2 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	woodId := args[0]
	woodFeatureCode :=args[1]
	if woodId == "" || woodFeatureCode == ""{
		return shim.Error("invalid args")
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	assetBytes, err := stub.GetState(constructWoodKey(assetId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("wood not found")
	}

	return shim.Success(assetBytes)
}

// 资产变更历史查询
func queryWoodHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 套路1：检查参数的个数
	if len(args) != 3 && len(args) != 2 {
		return shim.Error("not enough args")
	}

	// 套路2：验证参数的正确性
	woodId := args[0]
	woodFeatureCode := args[1]
	if woodId == "" || woodFeatureCode == "" {
		return shim.Error("invalid args")
	}

	queryType := "all"
	if len(args) == 2 {
		queryType = args[1]
	}

	if queryType != "all" && queryType != "enroll" && queryType != "exchange" {
		return shim.Error(fmt.Sprintf("queryType unknown %s", queryType))
	}

	// 套路3：验证数据是否存在 应该存在 or 不应该存在
	assetBytes, err := stub.GetState(constructWoodKey(assetId))
	if err != nil || len(assetBytes) == 0 {
		return shim.Error("asset not found")
	}

	// 查询相关数据
	keys := make([]string, 0)
	keys = append(keys, assetId)
	switch queryType {
	case "enroll":
		keys = append(keys, originOwner)
	case "exchange", "all": // 不添加任何附件key
	default:
		return shim.Error(fmt.Sprintf("unsupport queryType: %s", queryType))
	}
	result, err := stub.GetStateByPartialCompositeKey("history", keys)
	if err != nil {
		return shim.Error(fmt.Sprintf("query history error: %s", err))
	}
	defer result.Close()

	histories := make([]*WoodHistory, 0)
	for result.HasNext() {
		historyVal, err := result.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("query error: %s", err))
		}

		history := new(WoodHistory)
		if err := json.Unmarshal(historyVal.GetValue(), history); err != nil {
			return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
		}

		// 过滤掉不是资产转让的记录
		if queryType == "exchange" && history.OriginOwnerId == originOwner {
			continue
		}

		histories = append(histories, history)
	}

	historiesBytes, err := json.Marshal(histories)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}

	return shim.Success(historiesBytes)
}

// Init is called during Instantiate transaction after the chaincode container
// has been established for the first time, allowing the chaincode to
// initialize its internal data
func (c *AssertsExchangeCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
func (c *AssertsExchangeCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// 2.将Invoke方法与业务的方法关联
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "userRegister":
		return userRegister(stub, args)
	case "userDestroy":
		return userDestroy(stub, args)
	case "woodEnroll":
		return woodEnroll(stub, args)
	case "woodExchange":
		return woodExchange(stub, args)
	case "queryUser":
		return queryUser(stub, args)
	case "queryWood":
		return queryWood(stub, args)
	case "queryWoodHistory":
		return queryWoodHistory(stub, args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function: %s", funcName))
	}
	// stub.SetEvent("name", []byte("data"))
}

func main() {
	err := shim.Start(new(AssertsExchangeCC))
	if err != nil {
		fmt.Printf("Error starting AssertsExchange chaincode: %s", err)
	}
}
