document.addEventListener('DOMContentLoaded', async () => {
    const content = document.getElementById('content');

    try {
        // 获取当前活动标签页
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

        if (!tab || !tab.url) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        const url = new URL(tab.url);
        
        // 检查是否在能力平台域名下
        if (!url.hostname.includes('10086.cn') && !url.hostname.includes('sd.10086.cn')) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        // 获取所有可能的 cookies
        let allCookies = [];

        // 获取当前域名的 cookie
        const currentCookies = await chrome.cookies.getAll({
            domain: url.hostname
        });
        allCookies.push(...currentCookies);

        // 尝试从父域名查找
        const parts = url.hostname.split('.');
        for (let i = 1; i < parts.length - 1; i++) {
            const parentDomain = parts.slice(i).join('.');
            const parentCookies = await chrome.cookies.getAll({
                domain: parentDomain
            });
            allCookies.push(...parentCookies);
        }

        // 去重
        const seen = new Set();
        allCookies = allCookies.filter(c => {
            const key = c.domain + c.name;
            if (seen.has(key)) return false;
            seen.add(key);
            return true;
        });

        // 查找关键 token
        const tokenCookie = allCookies.find(c => c.name === '#openPortal#token#');
        const verificationCookie = allCookies.find(c => c.name === 'openPortalVerificationCode');
        const serviceIdCookie = allCookies.find(c => c.name === 'openPortalServiceID');

        if (tokenCookie && tokenCookie.value) {
            const token = tokenCookie.value;
            const loginCommand = `sdp login ${token}`;
            const aiInstallPrompt = '请帮我配置山东能力平台 CLI 助手（sdp），安装完成后执行 sdp onboard 命令';
            const copyIcon = '📋';

            content.innerHTML = `
                <div class="hint">1. AI 安装指令（推荐）</div>
                <div class="command-box" id="aiInstallBox" title="点击复制">
                    ${escapeHtml(aiInstallPrompt)}
                    <span class="copy-icon">${copyIcon}</span>
                </div>
                <div class="success-tip" id="aiInstallTip" style="display:none;">✓ 已复制，请发送给 AI</div>

                <div class="hint" style="margin-top: 12px;">2. 登录命令（点击复制）</div>
                <div class="command-box" id="authBox" title="点击复制">
                    ${escapeHtml(loginCommand)}
                    <span class="copy-icon">${copyIcon}</span>
                </div>
                <div class="success-tip" id="authTip" style="display:none;">✓ 已复制，请在终端执行</div>

                <div class="promo-section">
                    <div class="promo-title">🚀 推荐给其他同学</div>
                    <div class="command-box" id="promoBox" title="点击复制">
                        ${escapeHtml(aiInstallPrompt)}
                        <span class="copy-icon">${copyIcon}</span>
                    </div>
                    <div class="success-tip" id="promoTip" style="display:none;">✓ 已复制</div>
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

            // 认证命令点击
            document.getElementById('authBox').addEventListener('click', () => {
                navigator.clipboard.writeText(loginCommand).then(() => {
                    const tip = document.getElementById('authTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            // 推广命令点击
            document.getElementById('promoBox').addEventListener('click', () => {
                navigator.clipboard.writeText(aiInstallPrompt).then(() => {
                    const tip = document.getElementById('promoTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

        } else {
            content.innerHTML = '<div class="error">未找到登录凭证<br>请先登录山东能力平台</div>';
        }

    } catch (error) {
        console.error(error);
        content.innerHTML = '<div class="error">获取登录凭证失败<br>请刷新页面后重试</div>';
    }
});

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
