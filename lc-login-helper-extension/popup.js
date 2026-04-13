document.addEventListener('DOMContentLoaded', async () => {
    const content = document.getElementById('content');

    try {
        // 获取当前活动标签页
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

        if (!tab || !tab.url) {
            content.innerHTML = '<div class="error">请在灵畿平台页面，点击本图标</div>';
            return;
        }

        const url = new URL(tab.url);

        // 获取所有可能的 cookies（从当前域名和所有父域名）
        let allCookies = [];

        // 首先获取当前域名的 cookie
        const currentCookies = await chrome.cookies.getAll({
            domain: url.hostname,
            name: 'MOSS_SESSION'
        });
        allCookies.push(...currentCookies);

        // 尝试从父域名查找
        const parts = url.hostname.split('.');
        for (let i = 1; i < parts.length - 1; i++) {
            const parentDomain = parts.slice(i).join('.');
            const parentCookies = await chrome.cookies.getAll({
                domain: parentDomain,
                name: 'MOSS_SESSION'
            });
            allCookies.push(...parentCookies);
        }

        // 去重（根据 domain 和 value）
        const seen = new Set();
        allCookies = allCookies.filter(c => {
            const key = c.domain + c.value;
            if (seen.has(key)) return false;
            seen.add(key);
            return true;
        });

        // 优先选择 .hq.cmcc 域名下的 cookie
        let cookie = allCookies.find(c => c.domain === '.hq.cmcc');

        // 如果没有精确匹配 .hq.cmcc，尝试以 .hq.cmcc 结尾的域名
        if (!cookie) {
            cookie = allCookies.find(c => c.domain.endsWith('.hq.cmcc'));
        }

        // 如果还是没有，使用第一个找到的
        if (!cookie && allCookies.length > 0) {
            cookie = allCookies[0];
        }

        if (cookie && cookie.value) {
            const lcCommand = `lc login ${cookie.value}`;
            const aiInstallPrompt = '请帮我全局安装或更新@lingji/lc，使用npm/pnpm/yarn均可，安装完成后执行 lc onboard 命令';
            const humanInstallCommand = 'npm install -g @lingji/lc';
            const copyIcon = '📋';

            content.innerHTML = `
                <div class="hint">1. AI 安装指令（推荐）</div>
                <div class="command-box" id="aiInstallBox" title="点击复制">
                    ${escapeHtml(aiInstallPrompt)}
                    <span class="copy-icon">${copyIcon}</span>
                </div>
                <div class="success-tip" id="aiInstallTip" style="display:none;">✓ 已复制，请发送给 AI</div>

                <div class="hint" style="margin-top: 12px;">2. 手动安装命令</div>
                <div style="margin-bottom: 8px;">
                    <span class="tab-btn active" data-mode="npm" id="tabNpm">npm</span>
                    <span class="tab-btn" data-mode="pnpm" id="tabPnpm">pnpm</span>
                    <span class="tab-btn" data-mode="yarn" id="tabYarn">yarn</span>
                </div>
                <div class="command-box" id="manualInstallBox" title="点击复制">
                    ${escapeHtml(humanInstallCommand)}
                    <span class="copy-icon">${copyIcon}</span>
                </div>
                <div class="success-tip" id="manualInstallTip" style="display:none;">✓ 已复制</div>

                <div class="hint" style="margin-top: 12px;">3. 认证命令(认证完成后，AI 助手可直接调用相关接口，实现自动化操作)</div>
                <div class="command-box" id="authBox" title="点击复制">
                    ${escapeHtml(lcCommand)}
                    <span class="copy-icon">${copyIcon}</span>
                </div>
                <div class="success-tip" id="authTip" style="display:none;">✓ 已复制，请在终端执行</div>

                <div class="promo-section">
                    <div class="promo-title">🚀 推荐给其他同学</div>
                    <div class="promo-box" id="promoBox" title="点击复制">
                         ${escapeHtml(aiInstallPrompt)}
                        <span class="copy-icon">${copyIcon}</span>
                    </div>
                    <div class="success-tip" id="promoTip" style="display:none;">✓ 已复制，可粘贴给同学使用</div>
                </div>
            `;

            // AI 安装指令点击
            document.getElementById('aiInstallBox').addEventListener('click', () => {
                navigator.clipboard.writeText(aiInstallPrompt).then(() => {
                    const tip = document.getElementById('aiInstallTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            // 标签切换
            const tabNpm = document.getElementById('tabNpm');
            const tabPnpm = document.getElementById('tabPnpm');
            const tabYarn = document.getElementById('tabYarn');
            const manualInstallBox = document.getElementById('manualInstallBox');

            tabNpm.addEventListener('click', () => {
                setActiveTab(tabNpm, [tabPnpm, tabYarn]);
                manualInstallBox.innerHTML = escapeHtml('npm install -g @lingji/lc') + '<span class="copy-icon">' + copyIcon + '</span>';
                document.getElementById('manualInstallTip').style.display = 'none';
            });

            tabPnpm.addEventListener('click', () => {
                setActiveTab(tabPnpm, [tabNpm, tabYarn]);
                manualInstallBox.innerHTML = escapeHtml('pnpm add -g @lingji/lc') + '<span class="copy-icon">' + copyIcon + '</span>';
                document.getElementById('manualInstallTip').style.display = 'none';
            });

            tabYarn.addEventListener('click', () => {
                setActiveTab(tabYarn, [tabNpm, tabPnpm]);
                manualInstallBox.innerHTML = escapeHtml('yarn global add @lingji/lc') + '<span class="copy-icon">' + copyIcon + '</span>';
                document.getElementById('manualInstallTip').style.display = 'none';
            });

            // 手动安装命令点击
            manualInstallBox.addEventListener('click', () => {
                const text = manualInstallBox.childNodes[0].textContent.trim();
                navigator.clipboard.writeText(text).then(() => {
                    const tip = document.getElementById('manualInstallTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            // 认证命令点击
            document.getElementById('authBox').addEventListener('click', () => {
                navigator.clipboard.writeText(lcCommand).then(() => {
                    const tip = document.getElementById('authTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            // 推广命令点击
            const promoText = '请帮我全局安装@lingji/lc，使用npm/pnpm/yarn均可，安装完成后执行 lc onboard 命令';
            document.getElementById('promoBox').addEventListener('click', () => {
                navigator.clipboard.writeText(promoText).then(() => {
                    const tip = document.getElementById('promoTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

        } else {
            content.innerHTML = '<div class="error">请在灵畿平台页面，点击本图标</div>';
        }

    } catch (error) {
        content.innerHTML = '<div class="error">请在灵畿平台页面，点击本图标</div>';
    }
});

function setActiveTab(active, others) {
    active.classList.add('active');
    others.forEach(tab => tab.classList.remove('active'));
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
