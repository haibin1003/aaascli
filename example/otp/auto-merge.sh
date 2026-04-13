#!/bin/bash
# 自动化合并脚本 - 配合 OTP 使用
# 使用方法: ./auto-merge.sh <mr-id>

set -e

MR_ID=$1

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [ -z "$MR_ID" ]; then
    echo -e "${RED}错误: 请提供 MR ID${NC}"
    echo "用法: $0 <mr-id>"
    exit 1
fi

# 检查 lc-otp-gen 是否安装
if ! command -v lc-otp-gen &> /dev/null; then
    echo -e "${RED}错误: lc-otp-gen 未安装${NC}"
    echo "请先运行: make install-otp-gen"
    exit 1
fi

# 检查 lc 是否安装
if ! command -v lc &> /dev/null; then
    echo -e "${RED}错误: lc 未安装${NC}"
    echo "请先运行: make install"
    exit 1
fi

echo -e "${YELLOW}正在获取 OTP 验证码...${NC}"

# 获取当前 OTP 验证码
CODE=$(lc-otp-gen code 2>/dev/null | grep "验证码:" | head -1 | sed 's/.*验证码: //')

if [ -z "$CODE" ] || [ ${#CODE} -ne 6 ]; then
    echo -e "${RED}错误: 获取 OTP 验证码失败${NC}"
    echo "请确保已配置 OTP 账户: lc-otp-gen list"
    exit 1
fi

echo -e "${GREEN}获取到 OTP 验证码: $CODE${NC}"

# 验证 OTP，创建会话
echo -e "${YELLOW}正在验证 OTP...${NC}"
if ! lc otp verify "$CODE" > /dev/null 2>&1; then
    echo -e "${RED}错误: OTP 验证失败${NC}"
    exit 1
fi

echo -e "${GREEN}OTP 验证成功，会话有效期 5 分钟${NC}"

# 执行合并
echo -e "${YELLOW}正在合并 MR #$MR_ID...${NC}"
lc pr merge "$MR_ID" --squash --delete-branch

echo -e "${GREEN}合并完成!${NC}"
