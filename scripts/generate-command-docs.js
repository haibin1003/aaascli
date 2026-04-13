#!/usr/bin/env node

/**
 * 命令文档生成脚本
 * 运行方式: node scripts/generate-command-docs.js
 *
 * 此脚本通过执行 lc --help 和各子命令的 --help，
 * 收集命令信息并输出 Markdown 格式的文档。
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const LC_BINARY = path.join(__dirname, '..', 'bin', 'lc');

// 检查二进制文件是否存在
function checkBinary() {
  if (!fs.existsSync(LC_BINARY)) {
    console.error(`错误: 未找到 lc 二进制文件: ${LC_BINARY}`);
    console.error('请先运行: make build');
    process.exit(1);
  }
}

// 执行命令并返回输出
function execCommand(cmd) {
  try {
    return execSync(cmd, { encoding: 'utf-8', cwd: path.join(__dirname, '..') });
  } catch (e) {
    return e.stdout || e.message;
  }
}

// 获取命令帮助输出
function getHelpOutput(args = []) {
  const cmd = `${LC_BINARY} ${args.join(' ')} --help`;
  return execCommand(cmd);
}

// 解析全局帮助获取子命令列表
function parseSubcommands(helpOutput) {
  const commands = [];
  const lines = helpOutput.split('\n');
  let inCommandsSection = false;

  for (const line of lines) {
    // 检测命令列表开始
    if (line.includes('Available Commands:') || line.includes('Commands:')) {
      inCommandsSection = true;
      continue;
    }

    // 检测命令列表结束（遇到空行或其他部分）
    if (inCommandsSection && line.trim() === '') {
      continue;
    }

    if (inCommandsSection && line.startsWith('  ')) {
      const match = line.match(/^\s+([a-z-]+)\s+/);
      if (match && !['help'].includes(match[1])) {
        commands.push(match[1]);
      }
    }

    // 遇到 Flags 部分结束命令收集
    if (inCommandsSection && (line.includes('Flags:') || line.includes('Global Flags:'))) {
      inCommandsSection = false;
    }
  }

  return [...new Set(commands)]; // 去重
}

// 解析帮助输出
function parseHelp(helpOutput) {
  const result = {
    usage: '',
    shortDesc: '',
    longDesc: '',
    flags: [],
    subcommands: []
  };

  const lines = helpOutput.split('\n');
  let section = 'header';

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // 提取 Usage
    if (line.startsWith('Usage:')) {
      result.usage = line.replace('Usage:', '').trim();
      continue;
    }

    // 提取描述（第一行非空且不是Usage）
    if (section === 'header' && !result.shortDesc && line.trim() && !line.startsWith('Usage:')) {
      result.shortDesc = line.trim();
      continue;
    }

    // 检测 Flags 部分
    if (line.includes('Flags:')) {
      section = 'flags';
      continue;
    }

    // 检测子命令部分
    if (line.includes('Available Commands:')) {
      section = 'commands';
      continue;
    }

    // 解析 Flags
    if (section === 'flags' && line.startsWith('  -')) {
      const flagMatch = line.match(/^\s+(-[a-zA-Z],?\s*)?(--[\w-]+)\s+(\S.*)$/);
      if (flagMatch) {
        result.flags.push({
          shorthand: flagMatch[1] ? flagMatch[1].replace(',', '').trim() : '',
          name: flagMatch[2],
          description: flagMatch[3].trim()
        });
      }
    }

    // 解析子命令
    if (section === 'commands' && line.startsWith('  ')) {
      const cmdMatch = line.match(/^\s+([a-z-]+)\s+(\S.*)$/);
      if (cmdMatch) {
        result.subcommands.push({
          name: cmdMatch[1],
          description: cmdMatch[2].trim()
        });
      }
    }
  }

  return result;
}

// 生成 Markdown 文档片段
function generateMarkdown() {
  console.log('# 灵畿 CLI 助手 - 命令大全\n');
  console.log('> **生成时间**:', new Date().toLocaleString(), '\n');
  console.log('---\n');

  // 获取根命令帮助
  const rootHelp = getHelpOutput([]);
  console.log('## 全局参数\n');
  console.log('```');
  console.log(rootHelp);
  console.log('```\n');

  // 获取子命令
  const subcommands = parseSubcommands(rootHelp);

  // 为每个子命令生成文档
  for (const cmd of subcommands) {
    console.log(`\n## ${cmd} 命令\n`);

    // 获取子命令帮助
    const cmdHelp = getHelpOutput([cmd]);
    const parsed = parseHelp(cmdHelp);

    console.log(`**描述**: ${parsed.shortDesc}\n`);

    if (parsed.usage) {
      console.log('**用法**:');
      console.log('```bash');
      console.log(parsed.usage);
      console.log('```\n');
    }

    if (parsed.flags.length > 0) {
      console.log('**参数**:\n');
      console.log('| 参数 | 短选项 | 说明 |');
      console.log('|------|--------|------|');
      for (const flag of parsed.flags) {
        const short = flag.shorthand ? `-\`${flag.shorthand}\`` : '-';
        console.log(`| \`${flag.name}\` | ${short} | ${flag.description} |`);
      }
      console.log('');
    }

    if (parsed.subcommands.length > 0) {
      console.log('**子命令**:\n');
      for (const sub of parsed.subcommands) {
        console.log(`- \`${sub.name}\` - ${sub.description}`);

        // 获取子子命令帮助
        const subHelp = getHelpOutput([cmd, sub.name]);
        const subParsed = parseHelp(subHelp);

        if (subParsed.flags.length > 0) {
          console.log('\n  参数:');
          for (const flag of subParsed.flags) {
            const short = flag.shorthand ? `-${flag.shorthand}` : '';
            console.log(`  - ${flag.name} ${short}: ${flag.description}`);
          }
        }
      }
      console.log('');
    }
  }
}

// 主函数
function main() {
  checkBinary();

  console.log('正在生成命令文档...\n');

  // 输出到控制台，可以重定向到文件
  generateMarkdown();

  console.log('\n---\n');
  console.log('文档生成完成！');
  console.log('提示: 运行 `node scripts/generate-command-docs.js > commands-generated.md` 保存到文件');
}

main();
