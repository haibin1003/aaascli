#!/usr/bin/env node

/**
 * 版本同步脚本
 * 从 git tag 获取版本号，更新到 packages/lc/package.json，再同步到所有平台包
 */

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const PACKAGES_DIR = path.join(__dirname, "..", "packages");

// 从 git tag 获取版本号
function getVersionFromGitTag() {
  try {
    // 尝试获取当前 HEAD 的精确 tag（必须完全匹配）
    let tag;
    try {
      tag = execSync("git describe --tags --exact-match HEAD", {
        encoding: "utf8",
        stdio: ["pipe", "pipe", "ignore"]
      }).trim();
    } catch (e) {
      // 如果没有精确匹配的 tag，尝试获取最近的 tag
      console.warn("警告: 当前 commit 没有精确的 tag，尝试获取最近的 tag...");
      const describeOutput = execSync("git describe --tags --always", {
        encoding: "utf8",
        stdio: ["pipe", "pipe", "ignore"]
      }).trim();

      // 解析 git describe 输出，提取版本号部分
      // 格式可能是: v0.1.4-1-gc07a480 或 0.1.4-1-gc07a480
      // 我们只取第一个 '-' 之前的部分
      const match = describeOutput.match(/^v?(\d+\.\d+\.\d+)/);
      if (match) {
        tag = match[1];
        console.warn(`  从 '${describeOutput}' 提取版本号: ${tag}`);
      } else {
        tag = describeOutput;
      }
    }

    // 如果 tag 以 v 开头，去掉 v
    if (tag.startsWith("v")) {
      return tag.slice(1);
    }
    return tag;
  } catch (err) {
    console.error("获取 git tag 失败:", err.message);
    process.exit(1);
  }
}

// 获取版本号
const version = getVersionFromGitTag();
console.log(`Git Tag 版本: ${version}`);

// 更新主包版本
const mainPkgPath = path.join(PACKAGES_DIR, "lc", "package.json");
const mainPkg = JSON.parse(fs.readFileSync(mainPkgPath, "utf8"));
const oldMainVersion = mainPkg.version;

if (oldMainVersion !== version) {
  mainPkg.version = version;
  fs.writeFileSync(mainPkgPath, JSON.stringify(mainPkg, null, 2) + "\n");
  console.log(`  ✓ lc/package.json: ${oldMainVersion} → ${version}`);
} else {
  console.log(`  ✓ lc/package.json: ${version} (无变化)`);
}

// 获取所有包目录
const dirs = fs.readdirSync(PACKAGES_DIR);

let updated = 1; // 主包已算一个

// 同步到其他包
for (const dir of dirs) {
  if (dir === "lc") continue; // 主包已处理

  const pkgPath = path.join(PACKAGES_DIR, dir, "package.json");

  if (!fs.existsSync(pkgPath)) {
    continue;
  }

  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));

  // 更新版本号
  const oldVersion = pkg.version;
  pkg.version = version;

  // 写回文件
  fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n");

  if (oldVersion !== version) {
    console.log(`  ✓ ${dir}: ${oldVersion} → ${version}`);
  } else {
    console.log(`  ✓ ${dir}: ${version} (无变化)`);
  }
  updated++;
}

// 更新主包的 optionalDependencies 版本
const mainPkgUpdated = JSON.parse(fs.readFileSync(mainPkgPath, "utf8"));
let depsUpdated = false;
if (mainPkgUpdated.optionalDependencies) {
  for (const dep of Object.keys(mainPkgUpdated.optionalDependencies)) {
    if (dep.startsWith("@lingji/lc-")) {
      if (mainPkgUpdated.optionalDependencies[dep] !== version) {
        mainPkgUpdated.optionalDependencies[dep] = version;
        depsUpdated = true;
      }
    }
  }
}

if (depsUpdated) {
  fs.writeFileSync(mainPkgPath, JSON.stringify(mainPkgUpdated, null, 2) + "\n");
  console.log(`  ✓ lc/optionalDependencies: 已更新为 ${version}`);
}

console.log(`\n共更新 ${updated} 个包`);
console.log(`版本号来源: git tag (v${version})`);
