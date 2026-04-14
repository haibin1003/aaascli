document.addEventListener('DOMContentLoaded', async () => {
    const content = document.getElementById('content');

    try {
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

        if (!tab || !tab.url) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        const url = new URL(tab.url);

        if (!url.hostname.includes('10086.cn') && !url.hostname.includes('sd.10086.cn')) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        let allCookies = [];
        const currentCookies = await chrome.cookies.getAll({ domain: url.hostname });
        allCookies.push(...currentCookies);

        const parts = url.hostname.split('.');
        for (let i = 1; i < parts.length - 1; i++) {
            const parentDomain = parts.slice(i).join('.');
            const parentCookies = await chrome.cookies.getAll({ domain: parentDomain });
            allCookies.push(...parentCookies);
        }

        const seen = new Set();
        allCookies = allCookies.filter(c => {
            const key = c.domain + c.name;
            if (seen.has(key)) return false;
            seen.add(key);
            return true;
        });

        const tokenCookie = allCookies.find(c => c.name === '#openPortal#token#');

        if (tokenCookie && tokenCookie.value) {
            const token = tokenCookie.value;
            const loginCommand = 'sdp login ' + token;
            const aiInstallPrompt = '请帮我配置山东能力平台 CLI 助手（sdp），安装完成后执行 sdp onboard 命令';
            const copyIcon = '\uD83D\uDCCB';

            content.innerHTML =
                '<div class="hint">1. AI 配置指令（推荐）</div>' +
                '<div class="command-box" id="aiInstallBox" title="点击复制">' +
                    escapeHtml(aiInstallPrompt) +
                    '<span class="copy-icon">' + copyIcon + '</span>' +
                '</div>' +
                '<div class="success-tip" id="aiInstallTip" style="display:none;">\u2713 已复制，请发送给 AI</div>' +

                '<div class="hint" style="margin-top: 12px;">2. 登录命令（点击复制）</div>' +
                '<div class="command-box" id="authBox" title="点击复制">' +
                    escapeHtml(loginCommand) +
                    '<span class="copy-icon">' + copyIcon + '</span>' +
                '</div>' +
                '<div class="success-tip" id="authTip" style="display:none;">\u2713 已复制，请在终端执行</div>' +

                '<div class="promo-section">' +
                    '<div class="promo-title">\uD83D\uDE80 推荐给其他同学</div>' +
                    '<div class="command-box" id="promoBox" title="点击复制">' +
                        escapeHtml(aiInstallPrompt) +
                        '<span class="copy-icon">' + copyIcon + '</span>' +
                    '</div>' +
                    '<div class="success-tip" id="promoTip" style="display:none;">\u2713 已复制</div>' +
                '</div>';

            document.getElementById('aiInstallBox').addEventListener('click', () => {
                navigator.clipboard.writeText(aiInstallPrompt).then(() => {
                    const tip = document.getElementById('aiInstallTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            document.getElementById('authBox').addEventListener('click', () => {
                navigator.clipboard.writeText(loginCommand).then(() => {
                    const tip = document.getElementById('authTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

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