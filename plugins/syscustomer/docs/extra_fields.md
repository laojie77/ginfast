# Extra字段属性定义文档

## 概述
`extra`字段用于存储客户的扩展属性信息，采用JSON格式存储，便于扩展和维护。

## 属性定义

| 属性键 | 属性值 | 显示名称 | 描述 |
|--------|--------|----------|------|
| `house` | `house` | 房 | 客户是否有房产 |
| `car` | `car` | 车 | 客户是否有车辆 |
| `company` | `company` | 公司 | 客户是否有公司 |
| `credit` | `credit` | 信用卡 | 客户是否有信用卡 |
| `insurance` | `insurance` | 保单 | 客户是否有保单 |
| `work` | `work` | 打卡工资 | 客户是否有打卡工资 |
| `fund` | `fund` | 公积金 | 客户是否有公积金 |
| `social` | `social` | 社保 | 客户是否有社保 |
| `tax` | `tax` | 营业税 | 客户是否有营业税 |

## 数据格式

### 前端数据格式
```javascript
{
  extra: ["house", "car", "company"] // 选中的属性数组
}
```

### 后端存储格式
```json
{
  "extra": "{\"house\":1,\"car\":1,\"company\":1,\"credit\":0,\"insurance\":0,\"work\":0,\"fund\":0,\"social\":0,\"tax\":0}"
}
```

## 使用位置

### 前端
- 文件：`ginfast-ui/src/plugins/syscustomer/views/syscustomer/syscustomerlist.vue`
- 常量定义：`EXTRA_PROPERTIES`, `ALL_EXTRA_PROPERTIES`, `EXTRA_PROPERTY_LABELS`

### 后端
- 文件：`ginfast/plugins/syscustomer/models/syscustomerparam.go`
- 常量定义：`ExtraProperty*` 常量, `AllExtraProperties`, `ExtraPropertyLabels`

## 扩展说明

1. **添加新属性**：
   - 在前端和后端常量定义中添加新属性
   - 更新此文档
   - 无需修改数据库结构

2. **修改属性名称**：
   - 只需修改常量定义和显示名称
   - 保持属性键不变以确保数据兼容性

3. **数据迁移**：
   - 如需修改现有数据格式，需要编写数据迁移脚本

## 维护说明

- 所有属性定义应保持前后端一致
- 修改属性定义时需同步更新此文档
- 新增属性时需考虑数据兼容性