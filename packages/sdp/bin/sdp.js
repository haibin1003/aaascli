#!/usr/bin/env node

const { spawn } = require("child_process");
const fs = require("fs");
const path = require("path");

const platformPackages = {
  "linux-x64": "@aaas/sd-linux-x64",
  "linux-arm64": "@aaas/sd-linux-arm64",
  "darwin-x64": "@aaas/sd-darwin-x64",
  "darwin-arm64": "@aaas/sd-darwin-arm64",
  "win32-x64": "@aaas/sd-windows-x64",
  "win32-arm64": "@aaas/sd-windows-arm64",
};

function getBinaryPath() {
  const key = `${process.platform}-${process.arch}`;
  const pkgName = platformPackages[key];

  if (!pkgName) {
    console.error(`Unsupported platform: ${key}`);
    console.error("Supported platforms: " + Object.keys(platformPackages).join(", "));
    process.exit(1);
  }

  try {
    return require(pkgName);
  } catch (e) {
    console.error(`Failed to load platform package: ${pkgName}`);
    console.error("Please try installing sdp globally with: npm install -g @aaas/sd");
    process.exit(1);
  }
}

function copyDir(src, dst) {
  fs.mkdirSync(dst, { recursive: true });
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const srcPath = path.join(src, entry.name);
    const dstPath = path.join(dst, entry.name);
    if (entry.isDirectory()) {
      copyDir(srcPath, dstPath);
    } else {
      fs.copyFileSync(srcPath, dstPath);
    }
  }
}

function getDesktopPath() {
  const homeDir = require("os").homedir();
  const candidates = [path.join(homeDir, "Desktop"), path.join(homeDir, "桌面")];
  for (const p of candidates) {
    if (fs.existsSync(p)) return p;
  }
  return homeDir;
}

// Intercept "helper extract" to copy bundled browser extension
const args = process.argv.slice(2);
if (args.length >= 2 && args[0] === "helper" && args[1] === "extract") {
  const helperDir = path.join(__dirname, "..", "sdp-login-helper");
  if (fs.existsSync(helperDir)) {
    let outputDir = "";
    const outIdx = args.indexOf("-o");
    if (outIdx >= 0 && args[outIdx + 1]) {
      outputDir = args[outIdx + 1];
    } else if (args[2] && !args[2].startsWith("-")) {
      outputDir = args[2];
    }
    if (!outputDir) {
      outputDir = getDesktopPath();
    }
    const absPath = path.resolve(outputDir);
    const extensionDir = path.join(absPath, "sdp-login-helper");

    if (fs.existsSync(extensionDir)) {
      fs.rmSync(extensionDir, { recursive: true, force: true });
    }

    copyDir(helperDir, extensionDir);

    console.log(`浏览器扩展已释放到: ${extensionDir}`);
    console.log("\n安装步骤:");
    console.log("1. 打开 Chrome 浏览器，输入 chrome://extensions/");
    console.log("2. 开启右上角的「开发者模式」);
    console.log("3. 点击「加载已解压的扩展程序」);
    console.log("4. 选择上述目录");
    console.log("5. 登录平台后点击插件图标，复制 sdp login 命令");
    process.exit(0);
  }
}

const binaryPath = getBinaryPath();
const child = spawn(binaryPath, args, {
  stdio: "inherit",
  windowsHide: true,
});

child.on("exit", (code) => {
  process.exitCode = code || 0;
});
