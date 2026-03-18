# 渠道字段映射配置使用文档

## 一、功能概述

字段映射配置用于管理渠道数据回传到不同客户时的字段转换规则。当不同客户对接时，字段名称、星级计算规则可能不同，通过字段映射可以灵活配置，无需修改代码。

**回传接口地址**：`POST http://your-domain/api/callback/{channelKey}`

---

## 二、核心概念

### 2.1 映射类型

| 映射类型 | 说明 | 使用场景 |
|---------|------|---------|
| `direct` | 直接映射 | 字段名不同，值不变 |
| `static` | 静态值 | 固定返回某个值 |
| `transform` | 转换映射 | 值需要转换计算 |

### 2.2 字段说明

| 字段 | 说明 |
|------|------|
| `sourceField` | 源字段（系统内部字段名） |
| `targetField` | 目标字段（客户要求的字段名） |
| `mappingType` | 映射类型 |
| `staticValue` | 静态值（mappingType=static时使用） |
| `transformRules` | 转换规则（mappingType=transform时使用） |

---

## 三、转换规则语法

### 3.1 规则格式

```
值1:映射值1,值2:映射值2,default:默认值
```

### 3.2 特殊关键字

- `default`：默认值，当没有匹配到任何规则时使用
- `keep`：保持原值不变

### 3.3 示例

```
# 示例1：星级映射（6→1, 1→2, 其他→4）
6:1,1:2,default:4

# 示例2：星级映射（6→2, 1或2→4, 其他→5）
6:2,1:4,2:4,default:5

# 示例3：星级映射（6→0, 其他保持原值）
6:0,default:keep
```

---

## 四、使用步骤

### 4.1 创建渠道

1. 进入【渠道管理】页面
2. 点击【新增数据】
3. 填写渠道基本信息：
   - 名称：渠道名称
   - 渠道码：唯一标识（用于回传接口）
   - 渠道公司名称：对外显示名称
   - 渠道密钥：系统自动生成
   - 成功返回码：渠道回传数据时，用于判断响应是否成功的返回码（如：0, 1, 200）
   - 成功码字段名：响应中表示成功状态的字段名（如：code, status）
   - 状态：启用/停用
4. 点击保存

### 4.2 配置公司字段映射

1. 在渠道列表中，点击【公司配置】
2. 点击【新增配置】
3. 填写配置信息：
   - **公司平台**：选择对应的客户（对应orgId）
   - **城市**：选择数据归属城市
   - **是否回传**：是否启用数据回传
   - **字段映射配置**：配置字段转换规则
4. 点击保存

---

## 五、字段映射配置示例

### 场景1：客户A（orgId: 2567）

**需求：**
- 手机号字段名：`phoneMd5`
- 星级字段名：`star`
- 星级映射：6→1, 1→2, 其他→4

**配置步骤：**

1. 点击【添加字段映射】
   - 源字段：`md5_mobile`
   - 目标字段：`phoneMd5`
   - 映射类型：`direct`

2. 再次点击【添加字段映射】
   - 源字段：`customer_star`
   - 目标字段：`star`
   - 映射类型：`transform`
   - 转换规则：`6:1,1:2,default:4`

3. 保存配置

**最终生成的JSON：**
```json
[
  {
    "sourceField": "md5_mobile",
    "targetField": "phoneMd5",
    "mappingType": "direct"
  },
  {
    "sourceField": "customer_star",
    "targetField": "star",
    "mappingType": "transform",
    "transformRules": "6:1,1:2,default:4"
  }
]
```

**回传数据示例：**
```json
{
  "orgId": "2567",
  "phoneMd5": "abc123...",
  "star": 1
}
```

---

### 场景2：客户B（orgId: 2901）

**需求：**
- 手机号字段名：`phoneMd5`
- 星级字段名：`dataStandard`
- 星级映射：6→2, 1或2→4, 其他→5

**配置步骤：**

1. 添加手机号映射
   - 源字段：`md5_mobile`
   - 目标字段：`phoneMd5`
   - 映射类型：`direct`

2. 添加星级映射
   - 源字段：`customer_star`
   - 目标字段：`dataStandard`
   - 映射类型：`transform`
   - 转换规则：`6:2,1:4,2:4,default:5`

**最终生成的JSON：**
```json
[
  {
    "sourceField": "md5_mobile",
    "targetField": "phoneMd5",
    "mappingType": "direct"
  },
  {
    "sourceField": "customer_star",
    "targetField": "dataStandard",
    "mappingType": "transform",
    "transformRules": "6:2,1:4,2:4,default:5"
  }
]
```

**回传数据示例：**
```json
{
  "orgId": "2901",
  "phoneMd5": "abc123...",
  "dataStandard": 4
}
```

---

### 场景3：客户C（orgId: 1539）

**需求：**
- 手机号字段名：`phoneMd5`
- 星级字段名：`star`
- 星级映射：6→0, 其他保持原值

**配置步骤：**

1. 添加手机号映射
   - 源字段：`md5_mobile`
   - 目标字段：`phoneMd5`
   - 映射类型：`direct`

2. 添加星级映射
   - 源字段：`customer_star`
   - 目标字段：`star`
   - 映射类型：`transform`
   - 转换规则：`6:0,default:keep`

**最终生成的JSON：**
```json
[
  {
    "sourceField": "md5_mobile",
    "targetField": "phoneMd5",
    "mappingType": "direct"
  },
  {
    "sourceField": "customer_star",
    "targetField": "star",
    "mappingType": "transform",
    "transformRules": "6:0,default:keep"
  }
]
```

**回传数据示例：**
```json
{
  "orgId": "1539",
  "phoneMd5": "abc123...",
  "star": 3
}
```

---

### 场景4：使用 static 静态值

**需求：**
- 所有回传数据都需要带上数据来源标识 `source = "ginfast"`

**配置步骤：**

1. 添加静态值映射
   - 源字段：`any`（任意填，不会被使用）
   - 目标字段：`source`
   - 映射类型：`static`
   - 静态值：`ginfast`

**最终生成的JSON：**
```json
{
  "sourceField": "any",
  "targetField": "source",
  "mappingType": "static",
  "staticValue": "ginfast"
}
```

**回传数据示例：**
```json
{
  "orgId": "2567",
  "phoneMd5": "abc123...",
  "star": 1,
  "source": "ginfast"
}
```

---

### 场景5：配置 orgId 和回传地址

**需求：**
- 每个客户的 `orgId` 不同，需要通过字段映射配置
- 回传接口地址单独配置
- 成功返回码和响应码字段可配置（不同客户返回码不同）

**数据库表结构扩展：**

```sql
-- 公司配置表扩展字段
ALTER TABLE sys_channel_company ADD COLUMN callback_url VARCHAR(500) COMMENT '回传接口地址';
ALTER TABLE sys_channel_company ADD COLUMN field_mappings TEXT COMMENT '字段映射配置JSON（包含orgId、成功码等）';
```

**配置示例：**

客户A（orgId: 2567）：
- 回传地址：`https://customer-a.com/api/callback`
- 成功码：`200`
- 响应码字段：`code`

客户B（orgId: 2901）：
- 回传地址：`https://customer-b.com/receive`
- 成功码：`1`
- 响应码字段：`status`

**字段映射配置（包含orgId、成功码、响应码字段）：**
```json
{
  "mappings": [
    {
      "sourceField": "any",
      "targetField": "orgId",
      "mappingType": "static",
      "staticValue": "2567"
    },
    {
      "sourceField": "md5_mobile",
      "targetField": "phoneMd5",
      "mappingType": "direct"
    },
    {
      "sourceField": "customer_star",
      "targetField": "star",
      "mappingType": "transform",
      "transformRules": "6:1,1:2,default:4"
    }
  ],
  "successCode": "200",
  "responseCodeField": "code"
}
```

---

## 六、Go 后端实现

### 6.1 数据结构定义

```go
package models

// FieldMapping 字段映射配置
type FieldMapping struct {
    SourceField    string `json:"sourceField"`              // 源字段
    TargetField    string `json:"targetField"`              // 目标字段
    MappingType    string `json:"mappingType"`              // 映射类型: direct/static/transform
    StaticValue    string `json:"staticValue,omitempty"`    // 静态值
    TransformRules string `json:"transformRules,omitempty"` // 转换规则
}

// SysChannel 渠道表（包含成功码配置）
type SysChannel struct {
    // ... 原有字段 ...
    SuccessCode      string `gorm:"column:success_code;default:'0'" json:"successCode"`    // 成功返回码
    SuccessCodeField string `gorm:"column:success_code_field;default:'code'" json:"successCodeField"` // 成功码字段名
}

// SysChannelCompany 公司配置表扩展
type SysChannelCompany struct {
    // ... 原有字段 ...
    CallbackUrl   string `gorm:"column:callback_url" json:"callbackUrl"`     // 回传接口地址
    FieldMappings string `gorm:"column:field_mappings" json:"fieldMappings"` // 字段映射配置JSON
}

// FieldMappingConfig 完整的字段映射配置
type FieldMappingConfig struct {
    Mappings []FieldMapping `json:"mappings"` // 字段映射列表
}
```

### 6.2 字段映射服务

```go
package service

import (
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    
    "gin-fast/plugins/syschannel/models"
)

// FieldMappingService 字段映射服务
type FieldMappingService struct{}

// NewFieldMappingService 创建字段映射服务
func NewFieldMappingService() *FieldMappingService {
    return &FieldMappingService{}
}

// ApplyFieldMappings 应用字段映射
func (s *FieldMappingService) ApplyFieldMappings(
    sourceData map[string]interface{}, 
    mappings []models.FieldMapping,
) map[string]interface{} {
    result := make(map[string]interface{})
    
    for _, mapping := range mappings {
        switch mapping.MappingType {
        case "direct":
            // 直接映射：原样传递值
            if val, ok := sourceData[mapping.SourceField]; ok {
                result[mapping.TargetField] = val
            }
            
        case "static":
            // 静态值：使用配置的固定值
            result[mapping.TargetField] = mapping.StaticValue
            
        case "transform":
            // 转换映射：根据规则转换值
            sourceVal := sourceData[mapping.SourceField]
            transformedVal := s.transformValue(sourceVal, mapping.TransformRules)
            result[mapping.TargetField] = transformedVal
        }
    }
    
    return result
}

// transformValue 转换值
func (s *FieldMappingService) transformValue(value interface{}, rules string) interface{} {
    if rules == "" {
        return value
    }
    
    // 解析规则
    ruleMap := make(map[string]string)
    var defaultVal string
    
    pairs := strings.Split(rules, ",")
    for _, pair := range pairs {
        kv := strings.SplitN(pair, ":", 2)
        if len(kv) != 2 {
            continue
        }
        
        key := strings.TrimSpace(kv[0])
        val := strings.TrimSpace(kv[1])
        
        if key == "default" {
            defaultVal = val
        } else {
            ruleMap[key] = val
        }
    }
    
    // 将值转为字符串进行匹配
    valueStr := fmt.Sprintf("%v", value)
    
    // 查找匹配的规则
    if mappedVal, ok := ruleMap[valueStr]; ok {
        // 尝试转为int
        if intVal, err := strconv.Atoi(mappedVal); err == nil {
            return intVal
        }
        return mappedVal
    }
    
    // 使用默认值
    if defaultVal == "keep" {
        return value
    }
    
    if defaultVal != "" {
        // 尝试转为int
        if intVal, err := strconv.Atoi(defaultVal); err == nil {
            return intVal
        }
        return defaultVal
    }
    
    return value
}

// ParseFieldMappingConfig 解析完整的字段映射配置
func (s *FieldMappingService) ParseFieldMappingConfig(jsonStr string) (*models.FieldMappingConfig, error) {
    var config models.FieldMappingConfig
    if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
        return nil, err
    }
    // 设置默认值
    if config.SuccessCode == "" {
        config.SuccessCode = "0"
    }
    if config.ResponseCodeField == "" {
        config.ResponseCodeField = "code"
    }
    return &config, nil
}
```

### 6.3 回传接口实现

```go
package controller

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "gin-fast/plugins/syschannel/service"
)

// CallbackController 回传控制器
type CallbackController struct {
    channelService      *service.SysChannelService
    companyService      *service.SysChannelCompanyService
    fieldMappingService *service.FieldMappingService
}

// NewCallbackController 创建回传控制器
func NewCallbackController() *CallbackController {
    return &CallbackController{
        channelService:      service.NewSysChannelService(),
        companyService:      service.NewSysChannelCompanyService(),
        fieldMappingService: service.NewFieldMappingService(),
    }
}

// HandleCallback 处理回传请求
// POST /api/callback/:channelKey
func (c *CallbackController) HandleCallback(ctx *gin.Context) {
    channelKey := ctx.Param("channelKey")
    
    // 1. 根据channelKey获取渠道信息
    channel, err := c.channelService.GetByChannelKey(ctx, channelKey)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "渠道不存在"})
        return
    }
    
    // 2. 解析请求数据
    var sourceData map[string]interface{}
    if err := ctx.ShouldBindJSON(&sourceData); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
        return
    }
    
    // 3. 获取该渠道下的所有公司配置
    companyConfigs, err := c.companyService.GetByChannelId(ctx, channel.Id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取配置失败"})
        return
    }
    
    // 4. 遍历每个公司配置，应用字段映射并回传
    for _, config := range companyConfigs {
        // 跳过未启用回传的配置
        if config.IsStar != 1 {
            continue
        }
        
        // 跳过没有配置回传地址的
        if config.CallbackUrl == "" {
            continue
        }
        
        // 解析字段映射配置
        mappingConfig, err := c.fieldMappingService.ParseFieldMappingConfig(config.FieldMappings)
        if err != nil {
            continue
        }
        
        // 应用字段映射（包含orgId的配置）
        mappedData := c.fieldMappingService.ApplyFieldMappings(sourceData, mappingConfig.Mappings)

        // 5. 发送到客户接口（异步）
        go c.sendToCustomer(channel, config, mappedData)
    }

    ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success"})
}

// sendToCustomer 发送数据到客户接口
func (c *CallbackController) sendToCustomer(channel models.SysChannel, config models.SysChannelCompany, data map[string]interface{}) {
    // 1. 记录发送日志
    log.Printf("发送数据到客户 %d，地址: %s", config.TenantId, config.CallbackUrl)
    
    // 2. 序列化数据
    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Printf("序列化数据失败: %v", err)
        return
    }
    
    // 3. 发送HTTP POST请求
    resp, err := http.Post(
        config.CallbackUrl,
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        log.Printf("发送请求失败: %v", err)
        return
    }
    defer resp.Body.Close()
    
    // 4. 读取响应
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("读取响应失败: %v", err)
        return
    }
    
    // 5. 解析响应
    var response map[string]interface{}
    if err := json.Unmarshal(body, &response); err != nil {
        log.Printf("解析响应失败: %v", err)
        return
    }
    
    // 6. 验证响应是否成功（使用渠道级别的成功码配置）
    isSuccess := c.validateResponse(channel, response)

    // 7. 记录响应结果
    if isSuccess {
        log.Printf("回传成功，客户: %d", config.TenantId)
    } else {
        log.Printf("回传失败，客户: %d，响应: %s", config.TenantId, string(body))
    }
}

// validateResponse 验证响应是否成功
func (c *CallbackController) validateResponse(channel models.SysChannel, response map[string]interface{}) bool {
    // 获取响应码字段名（默认code）
    codeField := channel.SuccessCodeField
    if codeField == "" {
        codeField = "code"
    }

    // 获取期望的成功码
    expectedCode := channel.SuccessCode
    if expectedCode == "" {
        expectedCode = "0"
    }

    // 检查响应码是否匹配
    if val, ok := response[codeField]; ok {
        codeStr := fmt.Sprintf("%v", val)
        if codeStr == expectedCode {
            return true
        }
    }

    return false
}
```

### 6.4 使用示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "gin-fast/plugins/syschannel/service"
)

func main() {
    mappingService := service.NewFieldMappingService()
    
    // 模拟源数据
    sourceData := map[string]interface{}{
        "md5_mobile":    "e10adc3949ba59abbe56e057f20f883e",
        "customer_star": 6,
    }
    
    // 客户A的字段映射配置（包含orgId、手机号、星级映射）
    mappingsJSON := `{
        "mappings": [
            {"sourceField": "any", "targetField": "orgId", "mappingType": "static", "staticValue": "2567"},
            {"sourceField": "md5_mobile", "targetField": "phoneMd5", "mappingType": "direct"},
            {"sourceField": "customer_star", "targetField": "star", "mappingType": "transform", "transformRules": "6:1,1:2,default:4"}
        ]
    }`

    // 渠道配置（包含成功码配置）
    channel := models.SysChannel{
        SuccessCode:      "200",
        SuccessCodeField: "code",
    }
    
    mappingConfig, _ := mappingService.ParseFieldMappingConfig(mappingsJSON)
    result := mappingService.ApplyFieldMappings(sourceData, mappingConfig.Mappings)
    
    jsonBytes, _ := json.MarshalIndent(result, "", "  ")
    fmt.Println(string(jsonBytes))
    // 输出:
    // {
    //   "orgId": "2567",
    //   "phoneMd5": "e10adc3949ba59abbe56e057f20f883e",
    //   "star": 1
    // }
}
```

---

## 七、常见问题

### Q1：如何添加新的字段映射？

点击【添加字段映射】按钮，填写源字段、目标字段，选择映射类型，根据需要填写静态值或转换规则。

### Q2：如何删除字段映射？

点击对应映射行后面的【删除】按钮。

### Q3：转换规则写错了怎么办？

编辑配置，修改转换规则，保存即可生效。

### Q4：如何查看已配置的映射？

在公司配置列表的【字段映射】列，鼠标悬停可查看完整的映射配置。

### Q5：一个渠道可以配置多个公司吗？

可以，一个渠道可以配置多个公司，每个公司有自己独立的字段映射规则。

### Q6：static 类型如何使用？

当需要固定返回某个值时使用：
- 源字段：任意填写（不会被使用）
- 目标字段：客户要求的字段名
- 映射类型：选择 `static`
- 静态值：填写固定值

---

## 八、注意事项

1. **orgId 配置**：`orgId` 通过字段映射中的 `static` 类型配置，每个客户可以配置不同的值
2. **回传地址**：每个客户单独配置回传接口地址 `callback_url`
3. **成功码配置**：在渠道层面配置 `successCode` 和 `successCodeField`，统一控制该渠道下所有客户的成功码判断
4. **字段名一致性**：sourceField 必须与系统内部字段名一致
5. **规则格式正确**：转换规则格式为 `值:映射值`，多个规则用逗号分隔
6. **保存后生效**：配置修改后需要保存才能生效
7. **回传开关**：确保【是否回传】设置为启用状态，数据才会回传

---

## 九、附录：常用字段对照表

| 系统内部字段 | 说明 | 常用目标字段名 |
|-------------|------|--------------|
| `md5_mobile` | 手机号MD5 | `phoneMd5`, `mobileMd5` |
| `customer_star` | 客户类型/星级 | `star`, `dataStandard`, `level` |
| `name` | 姓名 | `userName`, `customerName` |
| `id_card` | 身份证号 | `idCard`, `idNumber` |
| `create_time` | 创建时间 | `createTime`, `applyTime` |

---

**文档版本**：v1.1  
**更新日期**：2026-03-10  
**适用系统**：渠道管理系统
