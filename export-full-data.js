// 在登录后的页面控制台执行此脚本，导出完整数据
(function() {
    console.log('=== 山东能力平台数据导出工具 ===\n');
    
    var result = {
        timestamp: new Date().toISOString(),
        url: window.location.href,
        title: document.title
    };
    
    // 1. 导出所有菜单
    console.log('【1. 菜单数据】');
    var allElements = document.querySelectorAll('*');
    var menuData = null;
    
    for (var i = 0; i < allElements.length; i++) {
        var vue = allElements[i].__vue__;
        if (vue && vue.$data && vue.$data.menuList) {
            menuData = vue.$data.menuList.map(function(item) {
                return {
                    menuName: item.menuName,
                    menuUrl: item.menuUrl,
                    menuId: item.menuId,
                    description: item.description
                };
            });
            break;
        }
    }
    result.menuList = menuData || [];
    console.log('菜单项:', result.menuList.length);
    result.menuList.forEach(function(m) {
        console.log('  - ' + m.menuName + ': ' + m.menuUrl);
    });
    
    // 2. 导出所有链接
    console.log('\n【2. 页面链接】');
    var allLinks = [];
    document.querySelectorAll('a').forEach(function(a) {
        var href = a.getAttribute('href') || '';
        var text = a.innerText || a.textContent || '';
        if (href && href.startsWith('/')) {
            allLinks.push({
                text: text.trim().substring(0, 50),
                href: href
            });
        }
    });
    // 去重
    var uniqueLinks = [];
    var seen = new Set();
    allLinks.forEach(function(l) {
        if (!seen.has(l.href)) {
            seen.add(l.href);
            uniqueLinks.push(l);
        }
    });
    result.links = uniqueLinks;
    console.log('链接数量:', uniqueLinks.length);
    uniqueLinks.slice(0, 20).forEach(function(l) {
        console.log('  - ' + l.text + ' -> ' + l.href);
    });
    
    // 3. 检查是否包含"数字服务"相关
    console.log('\n【3. 数字服务相关】');
    var digitalServiceLinks = uniqueLinks.filter(function(l) {
        return l.text.includes('数字') || 
               l.text.includes('服务') || 
               l.href.includes('digital') || 
               l.href.includes('service');
    });
    result.digitalServiceLinks = digitalServiceLinks;
    console.log('相关链接:', digitalServiceLinks.length);
    digitalServiceLinks.forEach(function(l) {
        console.log('  - ' + l.text + ' -> ' + l.href);
    });
    
    // 4. 导出当前用户的应用列表（如果有）
    console.log('\n【4. 查找应用数据】');
    var appData = null;
    for (var j = 0; j < allElements.length; j++) {
        var vue2 = allElements[j].__vue__;
        if (vue2 && vue2.$data) {
            var keys = Object.keys(vue2.$data);
            var appKeys = keys.filter(function(k) {
                return k.toLowerCase().includes('app') || 
                       k.toLowerCase().includes('application') ||
                       k.toLowerCase().includes('my') ||
                       k.toLowerCase().includes('order');
            });
            if (appKeys.length > 0) {
                appData = {
                    keys: appKeys,
                    sample: {}
                };
                appKeys.forEach(function(k) {
                    var val = vue2.$data[k];
                    if (Array.isArray(val)) {
                        appData.sample[k] = 'Array(' + val.length + ')';
                    } else if (typeof val === 'object') {
                        appData.sample[k] = Object.keys(val);
                    }
                });
                break;
            }
        }
    }
    result.appData = appData;
    console.log('应用数据:', appData ? '找到' : '未找到');
    
    // 5. 导出完整结果
    console.log('\n【5. 完整数据导出】');
    console.log(JSON.stringify(result, null, 2));
    
    // 复制到剪贴板
    navigator.clipboard.writeText(JSON.stringify(result, null, 2)).then(function() {
        console.log('\n✅ 已复制到剪贴板！');
    }).catch(function() {
        console.log('\n⚠️ 自动复制失败，请手动复制上面的 JSON');
    });
    
    return result;
})()
