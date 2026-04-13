// ============================================
// 脚本 3：导出我的应用/个人中心数据
// 在个人中心页面执行（尝试不同的 URL）
// ============================================

(function() {
    console.log('=== 个人中心/我的应用数据导出 ===\n');
    
    var result = {
        exportTime: new Date().toISOString(),
        pageUrl: window.location.href,
        pageTitle: document.title
    };
    
    // 1. 页面导航菜单
    console.log('【1. 个人中心导航菜单】');
    result.userMenu = [];
    document.querySelectorAll('.user-menu, .personal-menu, .menu-item, .nav-item').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('我的') || text.includes('应用') || text.includes('订单') || 
            text.includes('授权') || text.includes('审批')) {
            result.userMenu.push({
                name: text.trim(),
                href: el.getAttribute('href') || '',
                className: el.className
            });
        }
    });
    console.log('找到', result.userMenu.length, '个菜单项');
    result.userMenu.forEach(function(m) { console.log('  -', m.name); });
    
    // 2. 应用列表
    console.log('\n【2. 应用列表】');
    result.appList = [];
    var seenApps = new Set();
    
    document.querySelectorAll('.app-item, .application-item, .card, .el-card, [class*="app"]').forEach(function(el) {
        var titleEl = el.querySelector('.title, .name, h3, h4');
        if (titleEl) {
            var name = titleEl.innerText.trim();
            if (!seenApps.has(name)) {
                seenApps.add(name);
                
                // 查找状态
                var statusEl = el.querySelector('.status, [class*="status"]');
                var status = statusEl ? statusEl.innerText.trim() : '';
                
                // 查找操作按钮
                var buttons = [];
                el.querySelectorAll('button, a').forEach(function(btn) {
                    var text = btn.innerText || btn.textContent || '';
                    if (text.trim()) {
                        buttons.push(text.trim());
                    }
                });
                
                result.appList.push({
                    name: name,
                    status: status,
                    buttons: buttons
                });
            }
        }
    });
    
    console.log('找到', result.appList.length, '个应用');
    result.appList.slice(0, 5).forEach(function(app) {
        console.log('  -', app.name, app.status ? '(' + app.status + ')' : '');
    });
    
    // 3. 授权信息
    console.log('\n【3. 授权/权限信息】');
    result.authInfo = [];
    document.querySelectorAll('.auth-item, .permission-item, [class*="auth"]').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('授权') || text.includes('能力') || text.includes('服务')) {
            result.authInfo.push(text.trim().substring(0, 100));
        }
    });
    console.log('找到', result.authInfo.length, '条授权信息');
    
    // 4. 审批状态
    console.log('\n【4. 审批/订单状态】');
    result.approvalStatus = [];
    var bodyText = document.body.innerText;
    
    // 查找状态关键词
    var statusKeywords = ['审批中', '已通过', '已拒绝', '待审核', '已生效', '已过期', '处理中'];
    statusKeywords.forEach(function(keyword) {
        if (bodyText.includes(keyword)) {
            result.approvalStatus.push(keyword);
        }
    });
    console.log('找到状态:', result.approvalStatus.join(', ') || '无');
    
    // 5. Vue 数据
    console.log('\n【5. Vue 应用数据】');
    result.vueApps = [];
    var allElements = document.querySelectorAll('*');
    
    for (var i = 0; i < allElements.length && result.vueApps.length < 2; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data) {
            var keys = Object.keys(vue.$data);
            var appKeys = keys.filter(function(k) {
                var val = vue.$data[k];
                return (k.toLowerCase().includes('app') || 
                        k.toLowerCase().includes('application') ||
                        k.toLowerCase().includes('my')) &&
                       Array.isArray(val) && val.length > 0;
            });
            
            if (appKeys.length > 0) {
                var appData = {
                    elementIndex: i,
                    dataKeys: appKeys
                };
                
                appKeys.forEach(function(k) {
                    var arr = vue.$data[k];
                    appData[k] = {
                        count: arr.length,
                        sample: arr.slice(0, 2).map(function(item) {
                            return {
                                id: item.id || item.appId,
                                name: item.name || item.appName,
                                status: item.status
                            };
                        })
                    };
                });
                
                result.vueApps.push(appData);
            }
        }
    }
    console.log('找到', result.vueApps.length, '个 Vue 应用数据源');
    
    // 6. 操作按钮
    console.log('\n【6. 操作按钮】');
    result.actionButtons = [];
    document.querySelectorAll('button, a').forEach(function(el) {
        var text = el.innerText || el.textContent || '';
        if (text.includes('创建') || text.includes('新建') || text.includes('添加') ||
            text.includes('授权') || text.includes('管理') || text.includes('配置')) {
            result.actionButtons.push({
                text: text.trim(),
                tag: el.tagName,
                href: el.getAttribute('href') || ''
            });
        }
    });
    console.log('找到', result.actionButtons.length, '个操作按钮');
    result.actionButtons.forEach(function(b) { console.log('  -', b.text); });
    
    // 7. 导出结果
    console.log('\n=== 导出完成 ===');
    console.log(JSON.stringify(result, null, 2));
    
    navigator.clipboard.writeText(JSON.stringify(result, null, 2)).then(function() {
        console.log('\n✅ 已复制到剪贴板！请粘贴给我');
    }).catch(function() {
        console.log('\n⚠️ 自动复制失败，请手动复制上面的 JSON');
    });
    
    return result;
})();
