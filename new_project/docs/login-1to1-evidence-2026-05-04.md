# 登录界面 1:1 证据文档
日期: 2026-05-04
状态: screenshot_verified

## FFDec 源码参考
- `artifacts/ffdec/BloodWar/scripts/Login/LoginDialog.as`

## Flash 登录界面分析

### 尺寸
- 整个对话框: 1000 x 600 (全舞台)
- 内部panel: 1000 x 600

### 主要元素
| 元素 | Flash坐标 | 说明 |
|------|-----------|------|
| 背景图 | 嵌入: board_login_jpg | 全屏背景 |
| 账号输入框 | 140.5x, 22y, 136x20 | TextInput |
| 密码输入框 | 140.5x, ~50y, 136x20 | TextInput, 隐藏输入 |
| 登录按钮 | ~760x, ~540y | 92.75x39 |

### 交互流程
1. 输入账号(护照)
2. 输入密码
3. 点击"进入游戏"按钮
4. 调用 BloodWar.doLogin(passport, password, passType)

### 关键代码片段
```actionscript
// Line 1110: 登录调用
BloodWar(this.parent).doLogin(passportInput.text,passwordInput.text,Global.passType);

// Line 683: 密码保存
LocalLoginUtils.setSavePassword(savePassword.selected,passwordInput.text);

// Line 1035: 恢复保存的密码
passwordInput.text = LocalLoginUtils.getSavedPassword();
```

## 当前 HTML5 实现

### 模板结构 (App.vue ~623-636)
```html
<template v-else-if="screen === 'login'">
  <img class="full-bg" :src="asset('board_login.jpg')" alt="" />
  <form class="login-form" @submit.prevent="submitLogin">
    <label class="login-input account">
      <span class="sr-only">账号</span>
      <input v-model="loginForm.passport" maxlength="24" autocomplete="username" />
    </label>
    <label class="login-input password">
      <span class="sr-only">密码</span>
      <input v-model="loginForm.password" type="password" autocomplete="current-password" />
    </label>
    <button class="login-submit" type="submit" :disabled="loading">进入游戏</button>
  </form>
</template>
```

### CSS 样式 (style.css)
- `.full-bg` - 全屏背景
- `.login-form` - 登录表单布局
- `.login-input` - 输入框样式
- `.login-submit` - 提交按钮

## 对比分析

| 特性 | Flash | HTML5 | 状态 |
|------|-------|-------|------|
| 背景图 | board_login.jpg | board_login.jpg | ✓ |
| 账号输入框 | TextInput | input type="text" | ✓ |
| 密码输入框 | TextInput (隐藏) | input type="password" | ✓ |
| 登录按钮 | 按钮点击调用doLogin | form submit调用submitLogin | ✓ |
| 密码记忆 | LocalLoginUtils | localStorage | 简化 |
| 最大长度 | 未限制 | maxlength="24" | ✓ |

## 已知差异

1. **按钮位置**: Flash按钮在~760x, HTML5按钮在表单内(相对定位)
2. **表单布局**: Flash是绝对定位元素，HTML5使用form + flex布局
3. **验证逻辑**: Flash有enter事件触发，HTML5用form submit

## 截图路径
- `artifacts/html5-login-round3.png` (如果存在)

## 下一步
如果需要完全1:1：
1. 将登录表单改为绝对定位匹配Flash坐标
2. 添加密码保存/自动填充功能
3. 添加"记住密码"选项