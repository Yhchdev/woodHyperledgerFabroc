# 木材区块链系统

## 系统功能

* 用户注册登录（微信接口）
* 木材登记（属性、图片）：木材上链 用户绑定木材
* 木材交易记录：木材所有权的变更
* 查询功能：用户查询、木材查询、木材变更历史查询



## 业务实体

### 用户

* 名字
* 标识（身份证，电话号码，token）：考虑加密
* 木材资产列表



### 木材

* 编号
* 特征码
* 其他属性（重量、颜色、树种。。。）



### 木材交易记录

* 木材标识
* 原始拥有方（转让方）（木材登记 == null）
* 受让方



## 交互方法

### 一、用户开户

#### 参数

* 姓名
* 标识

### 二、用户销户

#### 参数

* 标识



### 三、木材登记

#### 参数（4个）

* 编号
* 图像特征值（先考虑一个，实际上会有多个位置取特征值）
* 原始图像（对应特征）：图片不存在区块链上
* 属性（json格式）
* 拥有者





### 四、木材交易

#### 参数

* 转让方
* 木材标识
* 受让方



### 五、用户查询

#### 参数

* 标识

#### 返回值

* 用户实体



### 六、木材查询

#### 参数

* 标识
* 特征值


#### 返回值
* 资产实体



### 七、木材交易历史记录查询

#### 参数

* 木材标识
* 特征值
* 交易类型（登记\交易\全部）（可选参数）

#### 返回值

* 交易列表









