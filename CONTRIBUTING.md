# 贡献指南

感谢你对CozyInsight项目感兴趣!我们欢迎任何形式的贡献。

---

## 🤝 如何贡献

### 报告Bug

如果发现Bug,请创建一个Issue并包含:

1. **问题描述** - 简洁清晰的描述
2. **复现步骤** - 详细的步骤
3. **预期行为** - 你期望发生什么
4. **实际行为** - 实际发生了什么
5. **环境信息** - OS, Go版本, Node版本等
6. **截图** - 如果适用

### 建议新功能

通过Issue提交功能建议:

1. **功能描述** - 清晰描述功能
2. **使用场景** - 为什么需要这个功能
3. **实现思路** - 如果有想法可以分享
4. **参考示例** - 如果有类似实现

### 提交代码

#### 1. Fork项目

```bash
# 访问GitHub页面点击Fork按钮
```

#### 2. 克隆到本地

```bash
git clone https://github.com/YOUR_USERNAME/CozyInsight.git
cd CozyInsight
```

#### 3. 创建分支

```bash
git checkout -b feature/your-feature-name
```

分支命名规范:
- `feature/xxx` - 新功能
- `fix/xxx` - Bug修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 代码重构
- `test/xxx` - 测试相关

#### 4. 开发

请遵循我们的编码规范(见下文)。

#### 5. 提交

使用语义化提交消息:

```bash
git commit -m "feat: add chart export feature"
```

提交类型:
- `feat` - 新功能
- `fix` - Bug修复
- `docs` - 文档
- `style` - 格式(不影响代码运行)
- `refactor` - 重构
- `test` - 测试
- `chore` - 构建/工具

#### 6. 推送

```bash
git push origin feature/your-feature-name
```

#### 7. 创建Pull Request

在GitHub上创建PR,填写:
- 标题要清晰
- 描述变更内容
- 关联相关Issue
- 添加截图(如果是UI改动)

---

## 📝 编码规范

### Go后端

#### 代码风格

```go
// ✅ 好的代码
func (s *userService) CreateUser(ctx context.Context, user *model.User) error {
    if user.Username == "" {
        return fmt.Errorf("username is required")
    }
    
    return s.repo.Create(ctx, user)
}

// ❌ 不好的代码
func CreateUser(u *model.User) {
    repo.Create(u) // 缺少context, 缺少错误处理
}
```

#### 规则

1. **总是传递context** - 第一个参数
2. **总是处理错误** - 不要忽略error
3. **不使用panic** - 使用error返回
4. **使用依赖注入** - 不用全局变量
5. **代码格式化** - 运行 `go fmt`
6. **添加注释** - 导出的函数必须注释
7. **编写测试** - 覆盖核心逻辑

### React前端

#### 代码风格

```typescript
// ✅ 好的代码
interface UserProps {
  userId: string;
  onUpdate: (user: User) => void;
}

const UserProfile: React.FC<UserProps> = ({ userId, onUpdate }) => {
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    loadUser(userId);
  }, [userId]);
  
  return <div>{/* ... */}</div>;
};

// ❌ 不好的代码  
const UserProfile = (props: any) => {
  // 缺少类型定义
}
```

#### 规则

1. **使用TypeScript** - 严格模式
2. **函数组件** - 用React.FC
3. **Hooks优先** - 不用Class组件
4. **Props类型** - 定义interface
5. **代码格式化** - Prettier
6. **组件拆分** - 保持组件小而简单

---

## 🧪 测试

### 后端测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./internal/service -v

# 测试覆盖率
go test -cover ./...
```

**要求**:
- 新功能必须有单元测试
- 测试覆盖率 > 70%
- 使用表驱动测试

### 前端测试

```bash
# 运行测试
npm test

# 测试覆盖率
npm run test:coverage
```

---

## 📖 文档

代码变更时更新相关文档:

- API变更 → 更新 `docs/API.md`
- 新功能 → 更新 README.md
- 配置变更 → 更新 `docs/DEVELOPMENT_GUIDE.md`

---

## 🔍 代码审查

所有PR都需要通过代码审查:

### 审查重点

1. **功能完整性** - 是否实现了预期功能
2. **代码质量** - 是否遵循编码规范
3. **测试覆盖** - 是否有足够测试
4. **性能影响** - 是否影响性能
5. **安全性** - 是否有安全隐患
6. **向后兼容** - 是否破坏现有功能

### 审查流程

1. 自动检查(CI)
2. 代码审查(至少1人)
3. 测试验证
4. 合并到main

---

## ✅ Pull Request清单

提交PR前请检查:

- [ ] 代码已格式化
- [ ] 通过所有测试
- [ ] 添加了新测试
- [ ] 更新了文档
- [ ] Commit消息符合规范
- [ ] 无冲突
- [ ] 通过CI检查

---

## 🎨 设计原则

### KISS原则

保持简单,避免过度设计。

### DRY原则

不要重复自己,复用代码。

### SOLID原则

- 单一职责
- 开闭原则
- 里氏替换
- 接口隔离
- 依赖倒置

---

## 💬 社区

- 💭 [讨论区](https://github.com/yourusername/CozyInsight/discussions)
- 🐛 [问题跟踪](https://github.com/yourusername/CozyInsight/issues)
- 📧 Email: dev@cozyinsight.com

---

## 📄 许可证

贡献的代码将使用 [Apache License 2.0](LICENSE) 许可。

---

**感谢你的贡献!** 🙏

每一个PR都让CozyInsight变得更好!
